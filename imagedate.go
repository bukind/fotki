package fotki

import (
	"fmt"
)

type ImageDate struct {
	year  int
	month int
	day   int
}

func (self ImageDate) IsEmpty() bool {
	return self.year == 0 && self.month == 0 && self.day == 0
}

func (self ImageDate) String() string {
	return self.DayString()
}

func (self ImageDate) DayString() string {
	return fmt.Sprintf("%04d-%02d-%02d", self.year, self.month, self.day)
}

func (self ImageDate) MonthString() string {
	return fmt.Sprintf("%04d-%02d", self.year, self.month)
}

func (self ImageDate) YearString() string {
	return fmt.Sprintf("%04d", self.year)
}

func (self ImageDate) Less(o ImageDate) bool {
	return self.year < o.year || (self.year == o.year &&
		(self.month < o.month ||
			(self.month == o.month && self.day < o.day)))
}

type ByImageDate []ImageDate

func (d ByImageDate) Len() int           { return len(d) }
func (d ByImageDate) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d ByImageDate) Less(i, j int) bool { return d[i].Less(d[j]) }
