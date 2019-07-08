package dnevnik76

import (
	"encoding/json"
	"net/http"
	"time"
)

type Client struct {
	Username    string       `json:"login"`
	Password    string       `json:"password"`
	SchoolID    int64        `json:"schoolId" xorm:"'school_id'"`
	Token       string       `json:"token"`
	http        *http.Client `xorm:"-"`
	currentInfo CurrentInfo  `xorm:"-"`
}

type CurrentInfo struct {
	//PersonID     int64  `json:"personId" xorm:"'person_id'"`
	SchoolID     int64  `json:"schoolId" xorm:"'school_id'"`
	ClassID      int64  `json:"clsId" xorm:"'class_id'"`
	ClassNumber  int    `json:"clsNum"`
	ClassChar    string `json:"clsChr"`
	EduYearStart int    `json:"eduYearStart"`
	EduYearEnd   int    `json:"eduYearEnd"`
}

type Region struct {
	ID   int64  `json:"id" xorm:"pk 'id'"`
	Name string `json:"name" xorm:"'name'"`
}

type School struct {
	ID       int64  `json:"id" xorm:"pk 'id'"`
	RegionID int64  `json:"regionId" xorm:"'region_id'"`
	Name     string `json:"name"`
	Type     string `json:"type"`
}

type Teacher struct {
	ID         int64  `json:"id" xorm:"pk autoincr 'id'"`
	UserID     string `json:"userId" xorm:"'user_id'"`
	SchoolID   int64  `json:"schoolId" xorm:"'school_id'"`
	FullName   string `json:"fullName"`
	CourseID   string `json:"courseId" xorm:"'course_id'"`
	CourseName string `json:"courseName"`
}

type Course struct {
	ID   int64  `json:"id" xorm:"pk 'id'"`
	Name string `json:"name"`
}

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

type Lperiod struct {
	SchoolID int64  `json:"schoolId" xorm:"'school_id'"`
	SYear    int    `json:"start_year"`
	EYear    int    `json:"end_year"`
	Name     string `json:"name"`
	Period   string `json:"period"`
}

func (p Lperiod) String() string {
	out, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return string(out)
}

type Mark struct {
	ID         int64     `json:"id" xorm:"pk autoincr 'id'"`
	UserID     string    `json:"userId" xorm:"'user_id'"`
	SchoolID   int64     `json:"schoolId" xorm:"'school_id'"`
	CourseID   int64     `json:"courseId" xorm:"'course_id'"`
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

type Message struct {
	ID       int64     `json:"id" xorm:"pk 'id'"`
	UserID   string    `json:"userId" xorm:"'user_id'"`
	Date     time.Time `json:"date"`
	From     string    `json:"from"`
	IsUnread bool      `json:"isUnread"`
	Subject  string    `json:"subject"`
	Body     string    `json:"body"`
}

type MarksListType int

const (
	Note MarksListType = iota
	List
	Date
)

func (s MarksListType) String() string {
	return [...]string{"note", "list", "date"}[s]
}

type MarkRange int

const (
	Month9 MarkRange = iota
	Month10
	Month11
	Month12
	Month1
	Month2
	Month3
	Month4
	Month5
	Month6
	Month7
	Month8
)

func (s MarkRange) String() string {
	return [...]string{"month9", "month10", "month11", "month12", "month1", "month2", "month3", "month4", "month5", "month6", "month7", "month8"}[s]
}
