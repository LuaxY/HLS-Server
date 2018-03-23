package errors

import "github.com/sirupsen/logrus"

type Error struct {
    Error  error
    Fields logrus.Fields
}
