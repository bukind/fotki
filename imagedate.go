package fotki

import (
    "fmt"
)

type ImageDate struct {
    year int
    month int
    day int
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
