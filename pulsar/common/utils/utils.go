package utils

import (
	"fmt"
	"strings"
)

func IsProducerNameExistsError(name string, err error) bool {
	return strings.Contains(err.Error(), fmt.Sprintf("Producer with name '%s' is already connected to topic", name))
}
