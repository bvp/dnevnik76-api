package dnevnik76

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"
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

func TestClient_Login(t *testing.T) {
	t.Log("Testing Login")
	err := client.Login()
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("client.GetCurrentInfo() - %#v", client.currentInfo)
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
		for _, school := range schools {
			sj, _ := json.Marshal(school)
			t.Logf(":: %s\n", string(sj))
		}
	}
}

func TestClient_GetMarks(t *testing.T) {
	marks, _ := client.GetMarksCurrent()
	t.Logf(":: size - %d", len(marks))
}

func TestClient_GetMarksPeriods(t *testing.T) {
	periods, _ := client.GetMarksPeriods()
	t.Logf(":: size - %d", len(periods))
	if DEBUG {
		for _, p := range periods {
			t.Logf(":: period - %s", p)
		}
	}
}

func TestClient_GetMarksNote(t *testing.T) {
	marks, _ := client.GetMarksForMonthWithType(Month5.String(), Note)
	t.Logf(":: size - %d", len(marks))
	for _, m := range marks {
		t.Logf(":: mark - %s", m)
	}
}

func TestClient_GetMarksList(t *testing.T) {
	marks, _ := client.GetMarksForMonthWithType(Month5.String(), List)
	t.Logf(":: size - %d", len(marks))
}

func TestClient_GetMarksFinal(t *testing.T) {
	marks, _ := client.GetMarksFinal()
	t.Logf(":: size - %d", len(marks))
	for _, m := range marks {
		t.Logf("%s", m)
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
			hwj, _ := json.Marshal(hw)
			t.Logf("  :: %s\n", string(hwj))
		}
	}
}

func TestClient_GetCourses(t *testing.T) {
	client.SetCookie("edu_year", "")
	client.GetMarksPeriods()
	courses, _ := client.GetCourses()
	t.Logf(":: size - %d", len(courses))

	client.SetCookie("edu_year", "2018")
	client.GetMarksPeriods()
	courses2018, _ := client.GetCourses()
	courses = append(courses, courses2018...)
	t.Logf(":: size for 2018 - %d", len(courses))

	client.SetCookie("edu_year", "2017")
	client.GetCurrentInfo()
	client.GetMarksPeriods()
	courses2017, _ := client.GetCourses()
	courses = append(courses, courses2017...)
	t.Logf(":: size for 2017 - %d", len(courses2017))

	client.SetCookie("edu_year", "2016")
	client.GetCurrentInfo()
	client.GetMarksPeriods()
	courses2016, _ := client.GetCourses()
	courses = append(courses, courses2016...)
	t.Logf(":: size for 2016 - %d", len(courses2016))

	courses = unique(courses)
	t.Logf(":: total size - %d", len(courses))
	if DEBUG {
		for _, course := range courses {
			cj, _ := json.Marshal(course)
			t.Logf("  :: %s\n", string(cj))
		}
	}
}

func TestClient_GetHomework2017(t *testing.T) {
	client.SetCookie("edu_year", "2017")
	hws, _ := client.GetHomework()
	t.Logf(":: size - %d", len(hws))
	for _, hw := range hws {
		hwj, _ := json.Marshal(hw)
		t.Logf("  :: %s\n", string(hwj))
	}
}

func TestClient_GetHomework2016(t *testing.T) {
	client.SetCookie("edu_year", "2016")
	hws, _ := client.GetHomework()
	t.Logf(":: size - %d", len(hws))
	for _, hw := range hws {
		hwj, _ := json.Marshal(hw)
		t.Logf("  :: %s\n", string(hwj))
	}
}

func TestClient_GetTeachers(t *testing.T) {
	teachers, err := client.GetTeachers()
	if err != nil {
		t.Logf("%s", err.Error())
	}
	t.Logf(":: size - %d", len(teachers))
	for _, teacher := range teachers {
		tj, _ := json.Marshal(teacher)
		t.Logf("  :: %s\n", string(tj))
	}
	client.SetCookie("items_perpage", "")
}

func setup() {
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
