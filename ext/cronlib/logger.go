package cronlib

import log "github.com/sirupsen/logrus"

type logWrapper struct {
	*log.Entry
}

func (c *logWrapper) Info(msg string, keysAndValues ...interface{}) {
	c.Entry.WithField("cron_msg", msg).Infof("%+v", keysAndValues)
}

func (c *logWrapper) Error(err error, msg string, keysAndValues ...interface{}) {
	c.Entry.WithError(err).WithField("cron_msg", msg).Errorf("%+v", keysAndValues)
}
