package auto

import "time"

type Option struct {
	data     string
	dateTime time.Time
	status   string
}

func (o *Option) Data() string {
	return o.data
}

func (o *Option) DateTime() time.Time {
	return o.dateTime
}

func (o *Option) Status() string {
	return o.status
}
