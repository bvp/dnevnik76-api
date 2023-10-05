// Package dnevnik76 model
package dnevnik76

import (
	"encoding/json"
	"net/http"
	"time"
)

// Client struct
type Client struct {
	Username    string       `json:"login"`
	Password    string       `json:"password"`
	RegionID    int64        `json:"region_id"`
	SchoolID    int64        `json:"schoolId" xorm:"'school_id'"`
	Token       string       `json:"token"`
	http        *http.Client `xorm:"-"`
	CurrentInfo CurrentInfo  `xorm:"-"`
}

// CurrentInfo struct
type CurrentInfo struct {
	//PersonID     int64  `json:"personId" xorm:"'person_id'"`
	RegionID     int64  `json:"regionId" xorm:"'region_id'"`
	SchoolID     int64  `json:"schoolId" xorm:"'school_id'"`
	SchoolName   string `json:"schoolName"`
	ClassID      int64  `json:"clsId" xorm:"'class_id'"`
	Class        string `json:"cls"`
	EduYearStart int    `json:"eduYearStart"`
	EduYearEnd   int    `json:"eduYearEnd"`
}

// Region struct
type Region struct {
	ID   int64  `json:"id" xorm:"pk 'id'"`
	Name string `json:"name" xorm:"'name'"`
}

// School struct
type School struct {
	ID       int64  `json:"id" xorm:"pk 'id'"`
	RegionID int64  `json:"regionId" xorm:"'region_id'"`
	Name     string `json:"name"`
	Type     string `json:"type"`
}

// Teacher struct
type Teacher struct {
	ID         int64  `json:"id" xorm:"pk autoincr 'id'"`
	UserID     string `json:"userId" xorm:"'user_id'"`
	SchoolID   int64  `json:"schoolId" xorm:"'school_id'"`
	FullName   string `json:"fullName"`
	CourseID   string `json:"courseId" xorm:"'course_id'"`
	CourseName string `json:"courseName"`
}

// Course struct
type Course struct {
	ID   int64  `json:"id" xorm:"pk 'id'"`
	Name string `json:"name"`
}

// Schedule struct
type Schedule struct {
	ID        int64     `json:"id" xorm:"pk autoincr 'id'"`
	SchoolID  int64     `json:"schoolId" xorm:"'school_id'"`
	StudentID int64     `json:"studentId" xorm:"'student_id'"`
	CourseID  int64     `json:"courseId" xorm:"'course_id'"`
	Subject   string    `json:"subject"`
	Homework  string    `json:"homework"`
	Marks     []int8    `json:"marks"`
	Date      time.Time `json:"date"`
}

// Homework struct
type Homework struct {
	ID         int64     `json:"id" xorm:"pk autoincr 'id'"`
	SchoolID   int64     `json:"schoolId" xorm:"'school_id'"`
	ClassID    int64     `json:"classId" xorm:"'class_id'"`
	Date       time.Time `json:"date"`
	DayOfWeek  string    `json:"dow"`
	CourseID   int64     `json:"courseId" xorm:"'course_id'"`
	CourseName string    `json:"courseName"`
	Homework   string    `json:"homework"`
	Subject    string    `json:"subject"`
}

// Lperiod struct
type Lperiod struct {
	SchoolID int64     `json:"schoolId" xorm:"'school_id'"`
	SYear    int       `json:"start_year"`
	EYear    int       `json:"end_year"`
	Name     string    `json:"name"`
	Period   string    `json:"period"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
}

func (p Lperiod) String() string {
	out, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return string(out)
}

// Mark struct
type Mark struct {
	ID         int64     `json:"id" xorm:"pk autoincr 'id'"`
	UserID     string    `json:"userId" xorm:"'user_id'"`
	SchoolID   int64     `json:"school_id" xorm:"'school_id'"`
	CourseID   int64     `json:"course_id" xorm:"'course_id'"`
	CourseName string    `json:"courseName"`
	Subject    string    `json:"subject"`
	HomeWork   string    `json:"homework"`
	Grade      []int8    `json:"grades"`
	DayOfWeek  string    `json:"dow"`
	Date       time.Time `json:"date"`
	SYear      int       `json:"s_year" xorm:"SMALLINT null"`
	EYear      int       `json:"e_year" xorm:"SMALLINT null"`
	Quarter    int       `json:"quarter" xorm:"SMALLINT null"`
	Annual     bool      `json:"annual" xorm:"null"`
}

func (m Mark) String() string {
	out, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	return string(out)
}

type MarksByDate []Mark

func (a MarksByDate) Len() int           { return len(a) }
func (a MarksByDate) Less(i, j int) bool { return a[i].Date.Before(a[j].Date) }
func (a MarksByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// Message struct
type Message struct {
	ID       int64     `json:"id" xorm:"pk 'id'"`
	UserID   string    `json:"userId" xorm:"'user_id'"`
	Date     time.Time `json:"date"`
	From     string    `json:"from"`
	IsUnread bool      `json:"isUnread"`
	Subject  string    `json:"subject"`
	Body     string    `json:"body"`
}

// MarksListType type
type MarksListType int

const (
	// Note type
	Note MarksListType = iota
	// List type
	List
	// Date type
	Date
)

func (s MarksListType) String() string {
	return [...]string{"note", "list", "date"}[s]
}

// MarkRange in month
type MarkRange int

const (
	// Month9 is September
	Month9 MarkRange = iota
	// Month10 is October
	Month10
	// Month11 is November
	Month11
	// Month12 is December
	Month12
	// Month1 is January
	Month1
	// Month2 is Febrary
	Month2
	// Month3 is March
	Month3
	// Month4 is April
	Month4
	// Month5 is May
	Month5
	// Month6 is June
	Month6
	// Month7 is July
	Month7
	// Month8 is August
	Month8
)

func (s MarkRange) String() string {
	return [...]string{"month9", "month10", "month11", "month12", "month1", "month2", "month3", "month4", "month5", "month6", "month7", "month8"}[s]
}
