package dnevnik76

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
)

var (
	client *Client
	cfg    config
)

type config struct {
	SchoolID  int64  `json:"school_id"`
	RegionID  int64  `json:"region_id"`
	MessageID int64  `json:"message_id"`
	Login     string `json:"login"`
	Password  string `json:"password"`
}

func arrayToString(a []int8, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
}

func TestClient_Login(t *testing.T) {
	t.Log("Testing Login")
	err := client.Login()
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("client.getCurrentInfo() - %#v", client.CurrentInfo)
}

func TestClient_GetRegions(t *testing.T) {
	regions, _ := GetRegions()
	t.Logf(":: size - %d", len(regions))
	if DEBUG {
		for _, r := range regions {
			t.Logf("  :: %d - %s", r.ID, r.Name)
		}
	}
}

func TestClient_GetSchool(t *testing.T) {
	schools, _ := GetSchools(cfg.RegionID)
	t.Logf(":: size - %d", len(schools))
	if DEBUG {
		for _, s := range schools {
			t.Logf("  :: %d - %s", s.ID, s.Name)
		}
	}
}

func TestClient_GetMarks(t *testing.T) {
	marks, _ := client.GetMarksCurrent()
	t.Logf(":: size - %d", len(marks))
	if DEBUG {
		for _, m := range marks {
			if m.Grade != nil {
				t.Logf(":: %s: %s - %s", m.Date.Format("2006.01.02"), m.CourseName, arrayToString(m.Grade, ","))
			}
		}
	}
}

func TestClient_GetMarksPeriods(t *testing.T) {
	periods, _ := client.GetMarksPeriods()
	t.Logf(":: size - %d", len(periods))
	if DEBUG {
		for _, p := range periods {
			t.Logf(":: %d-%d: %s - %s (%s - %s)", p.SYear, p.EYear, p.Name, p.Period, p.Start.Format("2006.01.02"), p.End.Format("2006.01.02"))
			inRange := dateWithinRange(time.Now(), p.Start, p.End)
			if inRange {
				if strings.Contains(p.Name, "четверть") {
					t.Logf("in range - %t: %s - %s (%s - %s)", inRange, p.Name, p.Period, p.Start.Format("2006.01.02"), p.End.Format("2006.01.02"))
				}
			}
		}
	}
}

func TestClient_GetMarksNote(t *testing.T) {
	marks, _ := client.GetMarksForWithType(Month1.String(), Note)
	t.Logf(":: size - %d", len(marks))
	if DEBUG {
		for _, m := range marks {
			t.Logf(":: %s: %s - %s", m.Date.Format("2006.01.02"), m.CourseName, arrayToString(m.Grade, ","))
		}
	}
}

func TestClient_GetMarksList(t *testing.T) {
	// marks, _ := client.GetMarksForWithType(fmt.Sprintf("month%d", time.Now().Month()), List)
	marks, _ := client.GetMarksForWithType(client.GetCurrentQuarter(), List)
	t.Logf(":: size - %d", len(marks))
	if DEBUG {
		sort.Sort(MarksByDate(marks))

		_marks := map[string][]string{}

		for _, m := range marks {
			t.Logf("%s: %s - %s", m.Date.Format("2006.01.02"), m.CourseName, arrayToString(m.Grade, ","))
			_marks[m.CourseName] = append(_marks[m.CourseName], arrayToString(m.Grade, ","))
		}

		for c, m := range _marks {
			t.Logf("%s: %s", c, m)
		}
	}
}

func TestClient_GetMarksFinal(t *testing.T) {
	marks, _ := client.GetMarksFinal()
	t.Logf(":: size - %d", len(marks))
	if DEBUG {
		for _, m := range marks {
			var q string
			switch m.Quarter {
			case 1:
				q = "1 четверть"
			case 2:
				q = "2 четверть"
			case 3:
				q = "3 четверть"
			case 4:
				q = "4 четверть"
			}
			if m.Annual {
				q = "Годовая"
			}
			t.Logf(":: %s (%d-%d): %s - %s", m.CourseName, m.SYear, m.EYear, q, arrayToString(m.Grade, ","))
		}
	}
}

func TestClient_GetMessagesCount(t *testing.T) {
	unread, total, _ := client.GetMessagesCount()
	t.Logf(":: unread: %d, total: %d", unread, total)
}

func TestClient_GetMessages(t *testing.T) {
	messages, _ := client.GetMessages()
	t.Logf(":: size - %d", len(messages))
}

func TestClient_GetMessage(t *testing.T) {
	m, _ := client.GetMessage(cfg.MessageID)
	mj, _ := json.Marshal(m)
	t.Logf(":: %s\n", string(mj))
}

func TestClient_GetHomework(t *testing.T) {
	client.SetCookie("edu_year", "")
	hws, _ := client.GetHomework()
	t.Logf(":: size - %d", len(hws))
	if DEBUG {
		for _, hw := range hws {
			t.Logf("  :: %s: %s - subject: %s, homework: %s", hw.Date.Format("2006.01.02"), hw.CourseName, hw.Subject, hw.Homework)
		}
	}
}

func TestClient_GetCourses(t *testing.T) {
	client.SetCookie("edu_year", "")
	client.GetMarksPeriods()
	courses, _ := client.GetCourses()
	t.Logf(":: size - %d", len(courses))

	years := []string{"2022", "2021", "2020", "2019", "2018"}
	for _, y := range years {
		client.SetCookie("edu_year", y)
		client.getCurrentInfo()
		client.GetMarksPeriods()
		coursesX, _ := client.GetCourses()
		courses = append(courses, coursesX...)
		t.Logf(":: size for %s - %d", y, len(courses))
	}

	courses = unique(courses)
	t.Logf(":: total size - %d", len(courses))
	if DEBUG {
		for _, course := range courses {
			t.Logf("  :: %s", course.Name)
		}
	}
}

func TestClient_GetHomework2022(t *testing.T) {
	client.SetCookie("edu_year", "2022")
	hws, _ := client.GetHomework()
	t.Logf(":: size - %d", len(hws))
	if DEBUG {
		for _, hw := range hws {
			t.Logf("  :: %s: %s - subject: %s, homework: %s", hw.Date.Format("2006.01.02"), hw.CourseName, hw.Subject, hw.Homework)
		}
	}
}

func TestClient_GetTeachers(t *testing.T) {
	client.SetCookie("edu_year", "")
	teachers, err := client.GetTeachers()
	if err != nil {
		t.Logf("ERR: %s", err.Error())
	}
	t.Logf(":: size - %d", len(teachers))
	for _, teacher := range teachers {
		t.Logf("  :: %s - %s", teacher.CourseName, teacher.FullName)
	}
	client.SetCookie("items_perpage", "")
}

func setup() {
	DEBUG = true
	file, _ := ioutil.ReadFile("config_test.json")
	cfg = config{}
	_ = json.Unmarshal([]byte(file), &cfg)

	client = NewClient(cfg.Login, cfg.Password, cfg.SchoolID, nil)
	err := client.Login()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func shutdown() {
	log.Println("shutdown")
}

func unique(intSlice []Course) []Course {
	keys := make(map[Course]bool)
	list := []Course{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}
