package helper

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/bwmarrin/snowflake"
)

var sn *snowflake.Node

func init() {
	_sn, err := snowflake.NewNode(rand.Int63n(1024)) // TODO: use a better way to generate a node id
	if err != nil {
		panic(err)
	}
	sn = _sn
}

func generateID() int64 {
	return sn.Generate().Int64()
}

func GenerateID(prefix string) string {
	return prefix + "_" + strconv.FormatInt(generateID(), 10)
}

func GetGenerator(seed int64) *snowflake.Node {
	node, err := snowflake.NewNode(seed)
	if err != nil {
		panic(err)
	}
	return node
}

type IDGenerator struct {
	seed   int64
	node   *snowflake.Node
	prefix string
}

func NewIDGenerator(seed int64, prefix string) *IDGenerator {
	if seed == 0 {
		seed = rand.Int63n(1024)
	}
	return &IDGenerator{
		seed:   seed,
		node:   GetGenerator(seed),
		prefix: prefix,
	}
}

func (g *IDGenerator) GenerateID(prefix ...string) string {

	numPart := strconv.FormatInt(g.node.Generate().Int64(), 10)

	parts := []string{}
	if g.prefix != "" {
		parts = append(parts, g.prefix)
	}

	for _, p := range prefix {
		if p != "" {
			parts = append(parts, p)
		}
	}
	joinedPrefix := ""
	if len(parts) > 0 {
		joinedPrefix = strings.Join(parts, "_") + "_"
	}
	return joinedPrefix + numPart
}
