package data

import (
	"checkerbox/internal/test"
	"time"

	"gorm.io/gorm"
)

type Report struct {
	gorm.Model
	Source        string
	Site          int
	OverallResult string
	ReportString  string
}

func NewReport() *Report {
	return &Report{}
}

func (r *Report) SetSource(source string) {
	r.Source = source
}

func (r *Report) SetSite(site int) {
	r.Site = site
}

func (r *Report) SetOverallResult(result test.ResultType) {
	r.OverallResult = result.String()
}

func (r *Report) AppendReportString(addition string) {
	dateString := time.Now().Format("15:4:5")
	r.ReportString += dateString + ": " + addition
}
