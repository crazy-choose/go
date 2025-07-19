package misc

import (
	"fmt"
	"time"
)

const (
	LayoutDay       = "20060102"
	LayoutSecond    = "20060102 15:04:05"
	LayoutMilSecond = "2006-01-02 15:04:05.000"
)

func DayFormat(t time.Time) string {
	return t.Format(LayoutDay)
}

func SecFormat(t time.Time) string {
	return t.Format(LayoutSecond)
}

func MilSecFormat(t time.Time) string {
	return t.Format(LayoutMilSecond)
}

func ZeroByDay(day string) *time.Time {
	parsedDate, err := time.Parse(LayoutDay, day)
	if err != nil {
		fmt.Println("解析日期出错:", err)
		return nil
	}

	// 获取日期的 0 点时间
	zt := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, parsedDate.Location())
	return &zt
}

func ZeroByTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
