// Package snowflake provides constructors and options for the snowflake library
// the library is utilized for generating unique ids
package snowflake

import (
	"github.com/bwmarrin/snowflake"
)

func NewSnowflake(
	opt *Options,
) (*snowflake.Node, error) {
	sf, err := snowflake.NewNode(opt.NodeNumber)
	if err != nil {
		return nil, err
	}

	return sf, nil
}
