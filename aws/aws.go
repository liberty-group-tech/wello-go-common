package aws

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/samber/lo"
)

type AwsConfig struct {
	Region string
}

type AwsService struct {
	secretsClient *secretsmanager.Client
	s3Client      *s3.Client
	presignClient *s3.PresignClient
}

func NewAwsService(cfg *AwsConfig) *AwsService {
	awsConfig, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(cfg.Region))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	s3Client := s3.NewFromConfig(awsConfig)
	presignClient := s3.NewPresignClient(s3Client)
	return &AwsService{
		secretsClient: secretsmanager.NewFromConfig(awsConfig),
		s3Client:      s3Client,
		presignClient: presignClient,
	}
}

func (s *AwsService) GetSecrets(secretArn *string) (map[string]string, error) {
	if secretArn == nil {
		return nil, nil
	}

	result, err := s.secretsClient.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: secretArn,
	})
	if err != nil {
		return nil, err
	}
	var secrets map[string]string
	if err := json.Unmarshal([]byte(*result.SecretString), &secrets); err != nil {
		return nil, err
	}
	return secrets, nil
}

func (s *AwsService) PutFile(bucket string, key string, file io.Reader) error {
	_, err := s.s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: lo.ToPtr(bucket),
		Key:    &key,
		Body:   file,
	})
	return err
}
func (s *AwsService) GetPresignedS3URL(bucket string, key string, expires *int) (string, error) {
	if expires == nil {
		expires = lo.ToPtr(1800) // 30 minutes in seconds
	}
	presignResult, err := s.presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: lo.ToPtr(bucket),
		Key:    &key,
	}, s3.WithPresignExpires(time.Duration(*expires)*time.Second))
	if err != nil {
		return "", err
	}
	return presignResult.URL, nil
}
