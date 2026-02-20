package services

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/pragma-proto/api/internal/models"
)

// PDFService generates PDF documents server-side.
// It uses HTML templates rendered to bytes, which can be converted to PDF
// via chromedp (headless Chrome) or a similar tool.
type PDFService struct{}

// NewPDFService creates a PDFService.
func NewPDFService() *PDFService {
	return &PDFService{}
}

// ReportCardData holds all the data needed to render a report card PDF.
type ReportCardData struct {
	School          *models.School
	Student         *models.Student
	StudentUser     *models.User
	AcademicPeriod  string
	GPA             float64
	CourseGrades    []CourseGradeRow
	TeacherComments string
	AdminComments   string
	GeneratedAt     time.Time
	IsFinalized     bool
}

// CourseGradeRow is one line in a report card.
type CourseGradeRow struct {
	CourseName   string
	TeacherName  string
	Percentage   float64
	LetterGrade  string
	Comment      string
}

// DocumentData holds data for generating enrollment certs / attendance letters.
type DocumentData struct {
	School           *models.School
	Student          *models.Student
	StudentUser      *models.User
	DocumentType     string
	VerificationCode string
	VerificationURL  string
	GeneratedAt      time.Time
	ExpiresAt        *time.Time
	CustomContent    string // for custom document type
	SignatoryName    string
	SignatoryTitle   string
}

// RenderReportCardHTML renders a report card as an HTML string.
// In production this would be fed to chromedp to produce a PDF.
func (s *PDFService) RenderReportCardHTML(data ReportCardData) ([]byte, error) {
	const tmpl = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Report Card — {{.Student.StudentNumber}}</title>
<style>
  body { font-family: Inter, sans-serif; font-size: 12pt; color: #1e293b; margin: 2cm; }
  .header { display: flex; align-items: center; margin-bottom: 24px; }
  .school-name { font-size: 20pt; font-weight: bold; }
  .student-info { margin-bottom: 20px; }
  table { width: 100%; border-collapse: collapse; margin-top: 16px; }
  th { background: #f1f5f9; text-align: left; padding: 8px; border: 1px solid #e2e8f0; }
  td { padding: 8px; border: 1px solid #e2e8f0; }
  .gpa { font-size: 16pt; font-weight: bold; margin-top: 16px; }
  .comments { margin-top: 24px; border-top: 1px solid #e2e8f0; padding-top: 16px; }
  .footer { margin-top: 40px; font-size: 9pt; color: #64748b; }
  .verification { font-size: 9pt; margin-top: 8px; }
</style>
</head>
<body>
<div class="header">
  {{if .School.LogoURL}}<img src="{{.School.LogoURL}}" height="60" style="margin-right:16px">{{end}}
  <div>
    <div class="school-name">{{.School.Name}}</div>
    <div>Report Card — {{.AcademicPeriod}}</div>
  </div>
</div>

<div class="student-info">
  <strong>Student:</strong> {{.StudentUser.FirstName}} {{.StudentUser.LastName}}<br>
  <strong>Student Number:</strong> {{.Student.StudentNumber}}<br>
  <strong>Grade Level:</strong> {{.Student.GradeLevel}}<br>
  <strong>Generated:</strong> {{.GeneratedAt.Format "January 2, 2006"}}
  {{if .IsFinalized}}<span style="color:#16a34a;font-weight:bold;"> — FINALIZED</span>{{end}}
</div>

<table>
  <thead>
    <tr>
      <th>Course</th>
      <th>Teacher</th>
      <th>%</th>
      <th>Grade</th>
      <th>Comment</th>
    </tr>
  </thead>
  <tbody>
  {{range .CourseGrades}}
    <tr>
      <td>{{.CourseName}}</td>
      <td>{{.TeacherName}}</td>
      <td>{{printf "%.1f" .Percentage}}%</td>
      <td><strong>{{.LetterGrade}}</strong></td>
      <td>{{.Comment}}</td>
    </tr>
  {{end}}
  </tbody>
</table>

<div class="gpa">Cumulative GPA: {{printf "%.3f" .GPA}}</div>

{{if .TeacherComments}}
<div class="comments">
  <strong>Teacher Comments</strong><br>{{.TeacherComments}}
</div>
{{end}}

{{if .AdminComments}}
<div class="comments">
  <strong>Administration Note</strong><br>{{.AdminComments}}
</div>
{{end}}

<div class="footer">
  <p>This document was generated on {{.GeneratedAt.Format "January 2, 2006 at 3:04 PM"}}.</p>
  {{if .School.Settings.SignatoryName}}
  <p>{{.School.Settings.SignatoryName}}, {{.School.Settings.SignatoryTitle}}</p>
  {{end}}
</div>
</body>
</html>`

	t, err := template.New("report_card").Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("pdf: parse report card template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("pdf: execute report card template: %w", err)
	}
	return buf.Bytes(), nil
}

// RenderDocumentHTML renders an official document (enrollment cert, etc.) as HTML.
func (s *PDFService) RenderDocumentHTML(data DocumentData) ([]byte, error) {
	const tmpl = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>{{.DocumentType}} — {{.School.Name}}</title>
<style>
  body { font-family: Inter, sans-serif; font-size: 12pt; color: #1e293b; margin: 3cm; }
  .school-header { text-align: center; margin-bottom: 40px; }
  .school-name { font-size: 20pt; font-weight: bold; }
  .title { font-size: 16pt; text-align: center; margin: 32px 0; font-weight: bold; text-transform: uppercase; letter-spacing: 0.05em; }
  .body { line-height: 1.8; margin: 24px 0; }
  .signature-block { margin-top: 60px; }
  .verification { margin-top: 40px; font-size: 9pt; color: #64748b; border-top: 1px solid #e2e8f0; padding-top: 16px; }
  .qr-hint { font-size: 8pt; }
</style>
</head>
<body>
<div class="school-header">
  {{if .School.LogoURL}}<img src="{{.School.LogoURL}}" height="80"><br>{{end}}
  <div class="school-name">{{.School.Name}}</div>
  {{if .School.Address}}<div>{{.School.Address}}</div>{{end}}
</div>

<div class="title">{{.DocumentType}}</div>

<div class="body">
  <p>This is to certify that <strong>{{.StudentUser.FirstName}} {{.StudentUser.LastName}}</strong>
  (Student Number: {{.Student.StudentNumber}}, Grade {{.Student.GradeLevel}})
  is {{if eq .Student.EnrollmentStatus "active"}}currently enrolled{{else}}{{.Student.EnrollmentStatus}}{{end}}
  at {{.School.Name}}.</p>

  {{if .CustomContent}}<p>{{.CustomContent}}</p>{{end}}

  <p>This document was issued on {{.GeneratedAt.Format "January 2, 2006"}}
  {{if .ExpiresAt}}and is valid through {{.ExpiresAt.Format "January 2, 2006"}}{{end}}.</p>
</div>

<div class="signature-block">
  <p>____________________________</p>
  {{if .SignatoryName}}<p>{{.SignatoryName}}<br>{{.SignatoryTitle}}</p>{{end}}
</div>

<div class="verification">
  <p><strong>Document Verification</strong></p>
  <p>To verify the authenticity of this document, visit:<br>
  <strong>{{.VerificationURL}}</strong></p>
  <p class="qr-hint">Verification Code: {{.VerificationCode}}</p>
</div>
</body>
</html>`

	t, err := template.New("document").Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("pdf: parse document template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("pdf: execute document template: %w", err)
	}
	return buf.Bytes(), nil
}
