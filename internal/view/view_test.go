package view

import (
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
)

func TestMain(m *testing.M) {
	testLogger, _ := test.NewNullLogger()
	logger = testLogger
}
