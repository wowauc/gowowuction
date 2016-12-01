package parser

import (
	"errors"
)

/*
parser.Provider
*/

type Provider interface {
	List() []string             // returns all basenames provided by this provider
	Get(string) ([]byte, error) // returns data loaded by basename
}

var ErrNotInList = errors.New("Entry not in list")
