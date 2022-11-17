// Package dnevnik76 parser
package dnevnik76

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"time"

	"net/http"
	"net/http/cookiejar"
	"net/url"

	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"

	"github.com/bvp/russiantime"
)

const (
	urlAjax         = "https://my.dnevnik76.ru/ajax"
	urlLogin        = "https://my.dnevnik76.ru/accounts/login/"
	urlHomework     = "https://my.dnevnik76.ru/homework/"
	urlMarksCurrent = "https://my.dnevnik76.ru/marks/current/"
	urlMarksFinal   = "https://my.dnevnik76.ru/marks/itog/"
	urlMessages     = "https://my.dnevnik76.ru/messages/input"
	urlTeachers     = "https://my.dnevnik76.ru/teachers/"

	sLoadSubjectsS = "loadSubjects('/ajax/subj/"
	sLoadSubjectsE = "', true)"
)

var (
	u *url.URL
	// DEBUG output
	DEBUG bool
)

// NewClient create new client
func NewClient(login string, password string, schoolID int64, httpClient *http.Client) *Client {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}

	ci := CurrentInfo{}
	ci.SchoolID = schoolID

	cookie := &http.Cookie{
		Name:   "items_perpage",
		Value:  "1000",
		Path:   "/",
		Domain: "my.dnevnik76.ru",
	}

	u, _ = url.Parse(urlLogin)
	cookies := jar.Cookies(u)
	cookies = append(cookies, cookie)
	jar.SetCookies(u, cookies)

	if httpClient == nil {
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Jar: jar,
		}
	}

	cli := &Client{
		Username:    login,
		Password:    password,
		SchoolID:    schoolID,
		http:        httpClient,
		CurrentInfo: ci,
	}

	return cli
}

// Login to dnevnik76.ru
func (cli *Client) Login() (err error) {
	resp, err := cli.http.Get(urlLogin)
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}
	cli.Token, _ = doc.Find(".login__form > input[name='csrfmiddlewaretoken']").First().Attr("value")

	payload := url.Values{
		"next":                {""}, // /marks/current/
		"csrfmiddlewaretoken": {cli.Token},
		"username":            {fmt.Sprintf("%s@%d", cli.Username, cli.SchoolID)},
		"fake_username":       {cli.Username},
		"password":            {cli.Password},
		"school":              {fmt.Sprintf("%d", cli.SchoolID)},
		"submit":              {""},
	}

	req, _ := http.NewRequest("POST", urlLogin, strings.NewReader(payload.Encode()))
	req.Header.Add("Referer", urlLogin)
	req.Header.Add("Host", "my.dnevnik76.ru")
	req.Header.Add("Origin", "https://my.dnevnik76.ru")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err = cli.http.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// doc, err = goquery.NewDocumentFromReader(resp.Body)
	// if err != nil {
	// 	return
	// }

	err = cli.getCurrentInfo()

	return
}

// getCurrentInfo for session
func (cli *Client) getCurrentInfo() (err error) {
	resp, err := cli.http.Get(urlHomework)
	if err != nil {
		return
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	cli.CurrentInfo.SchoolID = cli.SchoolID
	classNumber, classChar, err := getClassName(doc.Find("#auth_info > #role").Text())
	if err != nil {
		return
	}
	cli.CurrentInfo.ClassNumber = classNumber
	cli.CurrentInfo.ClassChar = classChar

	classIDText, _ := doc.Find("body").Attr("onload")
	if classIDText != "" {
		classIDText = strings.TrimSuffix(strings.TrimPrefix(classIDText, sLoadSubjectsS), sLoadSubjectsE)
	}
	classID, err := strconv.ParseInt(classIDText, 10, 32)
	cli.CurrentInfo.ClassID = classID

	var eys, eye int64
	eyr := strings.Split(strings.TrimSuffix(doc.Find("#eduyear > #curedy").Text(), " учебный год"), "-")
	eys, _ = strconv.ParseInt(eyr[0], 10, 32)
	eye, _ = strconv.ParseInt(eyr[1], 10, 32)
	cli.CurrentInfo.EduYearStart = int(eys)
	cli.CurrentInfo.EduYearEnd = int(eye)

	cli.ToJSON(cli.CurrentInfo)

	return
}

// ToJSON convert object to json notation
func (cli *Client) ToJSON(o interface{}) (result string) {
	oj, _ := json.Marshal(o)
	result = string(oj)
	return
}

// PrintCookies to print client cookies
func (cli *Client) PrintCookies() {
	log.Println(":: Cookies:")
	for _, cookie := range cli.http.Jar.Cookies(u) {
		log.Printf("  :: %s: %s\n", cookie.Name, cookie.Value)
	}
}

// SetCookie to set client cookie
func (cli *Client) SetCookie(name, value string) {
	cookie := &http.Cookie{
		Name:  name,
		Value: value,
		Path:  "/",
	}

	u, _ := url.Parse(urlLogin)
	var cookies []*http.Cookie
	cookies = append(cookies, cookie)
	cli.http.Jar.SetCookies(u, cookies)
	cli.getCurrentInfo()
}

// GetRegions to get client regions
func GetRegions() (regions []Region, err error) {
	resp, err := http.Get(fmt.Sprintf("%s/kladr/?login=true", urlAjax))
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	doc.Find("select > option").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Text())
		value, _ := s.Attr("value")
		regionID, _ := strconv.ParseInt(value, 10, 64)
		if regionID != 0 {
			regions = append(regions, Region{ID: regionID, Name: title})
		}
	})
	return
}

// GetSchools for selected region
func GetSchools(region int64) (schools []School, err error) {
	resp, err := http.Get(fmt.Sprintf("%s/school/%d/?login=true", urlAjax, region))
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}
	doc.Find("select > optgroup").Each(func(i int, s *goquery.Selection) {
		label, _ := s.Attr("label")
		s.Find("option").Each(func(i int, s2 *goquery.Selection) {
			title := strings.TrimSpace(s2.Text())
			value, _ := s2.Attr("value")
			schoolID, _ := strconv.ParseInt(value, 10, 64)
			schools = append(schools, School{ID: schoolID, RegionID: region, Name: title, Type: label})
		})
	})
	return
}

func dateWithinRange(date, start, end time.Time) bool {
	if date.After(start) && date.Before(end) {
		return true
	}
	return false
}

func (cli *Client) GetCurrentQuarter() (result string) {
	periods, _ := cli.GetMarksPeriods()
	for _, p := range periods {
		inRange := dateWithinRange(time.Now(), p.Start, p.End)
		if inRange {
			if strings.Contains(p.Name, "четверть") {
				result = p.Period
				break
			}
		}
	}
	return
}

// GetCourses to get subjects
func (cli *Client) GetCourses() (courses []Course, err error) {
	resp, err := cli.http.Get(fmt.Sprintf("%s/subj/%d", urlAjax, cli.CurrentInfo.ClassID))
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	doc.Find("select > option").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Text())
		value, _ := s.Attr("value")
		courseID, _ := strconv.ParseInt(value, 10, 64)
		if courseID != 0 {
			courses = append(courses, Course{ID: courseID, Name: title})
		}
	})
	return
}

// GetMarksPeriods to get marks periods
func (cli *Client) GetMarksPeriods() (periods []Lperiod, err error) {
	resp, err := cli.http.Get(urlMarksCurrent)
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}
	doc.Find("#mark_range > optgroup").Each(func(i int, s *goquery.Selection) {
		s.Find("option").Each(func(i int, s2 *goquery.Selection) {
			title := strings.TrimSpace(s2.Text())
			value, _ := s2.Attr("value")

			var resp *http.Response
			resp, err = cli.http.Get(fmt.Sprintf("%s%s/note", urlMarksCurrent, value))
			if err != nil {
				return
			}

			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				return
			}
			re := regexp.MustCompile(`(?P<start>(\d{1,2}\s[\p{L}]+\s\d{4}\sг\.)) по (?P<end>(\d{1,2}\s[\p{L}]+\s\d{4}\sг\.))`)
			n1 := re.SubexpNames()
			result := re.FindStringSubmatch(doc.Find("#content > h3").First().Text())
			m := map[string]string{}
			for i, n := range result {
				m[n1[i]] = n
			}

			period := Lperiod{
				SchoolID: cli.CurrentInfo.SchoolID,
				SYear:    cli.CurrentInfo.EduYearStart,
				EYear:    cli.CurrentInfo.EduYearEnd,
				Name:     title,
				Period:   value,
				Start:    russiantime.ParseDateString(m["start"]),
				End:      russiantime.ParseDateString(m["end"]),
			}
			periods = append(periods, period)
		})
	})
	return
}

// GetMarksCurrent to get marks for current month
func (cli *Client) GetMarksCurrent() (marks []Mark, err error) {
	return cli.GetMarksForWithType("", Note)
}

// GetMarksFor to get marks for specific month
func (cli *Client) GetMarksFor(p string) (marks []Mark, err error) {
	return cli.GetMarksForWithType(p, Note)
}

// GetMarksForWithType to get user marks
func (cli *Client) GetMarksForWithType(p string, t MarksListType) (marks []Mark, err error) {
	var sp string
	if p != "" {
		sp = fmt.Sprintf("%s/%s/", p, t.String())
	} else {
		sp = fmt.Sprintf("%s/", t.String())
	}
	var resp *http.Response
	resp, err = cli.http.Get(fmt.Sprintf("%s%s", urlMarksCurrent, sp))
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}
	rpt := regexp.MustCompile(`(\r\n)+|\r+|\n+|\t+|\s+`)
	log.Printf("page title - %s", rpt.ReplaceAllString(doc.Find("#content > h3").First().Text(), " "))
	switch t {
	case Note:
		doc.Find("#marks > div.week").Each(func(i int, s *goquery.Selection) {
			s.Find("div.dayofweek").Each(func(j int, s2 *goquery.Selection) {
				title := strings.TrimSpace(s2.Find("div.weekday > h3").First().Text())
				table := s2.Find("table")
				table.Find("tbody > tr").Each(func(k int, tr *goquery.Selection) {
					mark := Mark{}
					mark.SYear = cli.CurrentInfo.EduYearStart
					mark.EYear = cli.CurrentInfo.EduYearEnd
					mark.UserID = cli.Username
					mark.SchoolID = cli.SchoolID
					pd := strings.Split(strings.TrimRight(title, ")"), " (")
					mark.DayOfWeek = pd[0]
					mark.Date = russiantime.ParseDateString(pd[1])

					course := tr.Find("td:nth-child(1)").First().Text()
					mark.CourseName = course
					pt, _ := tr.Attr("title")
					lessonTitle := strings.TrimSpace(strings.TrimLeft(pt, "Тема: "))
					mark.Subject = lessonTitle
					hw := tr.Find("td:nth-child(2)").First().Text()
					mark.HomeWork = strings.TrimSpace(hw)
					tr.Find("td.col-mark > span.mark").Each(func(l int, m *goquery.Selection) {
						pm, _ := strconv.ParseInt(m.Text(), 10, 32)
						mark.Grade = append(mark.Grade, int8(pm))
					})
					marks = append(marks, mark)
				})
			})
		})
	case List:
		// TODO: fill DayOfWeek, Subject, HomeWork
		// URL: https://regex101.com/r/CwEys5/4
		doc.Find("#marks > #mark-row").Each(func(i int, s *goquery.Selection) {
			courseName := s.Find("div.mark-label").Text()
			s.Find("span.mark").Each(func(j int, sj *goquery.Selection) {
				if !sj.HasClass("avg") {
					mark := Mark{}
					mark.SYear = cli.CurrentInfo.EduYearStart
					mark.EYear = cli.CurrentInfo.EduYearEnd
					mark.CourseName = courseName
					el := sj.Find("a").First()
					onClick, _ := el.Attr("onclick")
					reg := regexp.MustCompile(`(showMarkInfo\(')(\d{1,2}\s\p{Cyrillic}*\s\d{4}\sг\.\s\(\p{Cyrillic}*\))?([^"]*)`)
					d := reg.ReplaceAllString(onClick, "${2}")
					pm, _ := strconv.ParseInt(el.Text(), 10, 32)
					mark.Date = russiantime.ParseDateString(d)
					mark.Grade = append(mark.Grade, int8(pm))
					marks = append(marks, mark)
				}
			})
		})
	case Date:
		log.Println("Not implemented right now")
	default:
		//
	}

	return
}

// GetMarksFinal to get final marks
func (cli *Client) GetMarksFinal() (marks []Mark, err error) {
	resp, err := cli.http.Get(urlMarksFinal)
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}
	courses, _ := cli.GetCourses()
	doc.Find("#marks > #wrap-col > #wrap-marks > div > #mark-row").Each(func(i int, s *goquery.Selection) {
		courseID, _ := s.Attr("name")

		s.Find(".mark").Each(func(j int, sj *goquery.Selection) {
			mark := Mark{}
			mark.SYear = cli.CurrentInfo.EduYearStart
			mark.EYear = cli.CurrentInfo.EduYearEnd
			mark.UserID = cli.Username
			mark.SchoolID = cli.SchoolID
			mark.CourseID, _ = strconv.ParseInt(courseID, 10, 32)
			for _, c := range courses {
				if c.ID == mark.CourseID {
					mark.CourseName = c.Name
					break
				}
			}
			data := func() (period string, fmark string) {
				el := sj.Find("a").First()
				onClick, _ := el.Attr("onclick")
				fmark = el.Text()
				reg := regexp.MustCompile(`(showMarkItogInfo\(')(\d\s\p{Cyrillic}*)?([^"]*)`)
				period = reg.ReplaceAllString(onClick, "${2}")
				return
			}
			if sj.HasClass("itg-q") {
				period, fmark := data()
				regd := regexp.MustCompile("[0-9]+")
				digs := regd.FindAllString(period, -1)
				if len(digs) > 0 {
					mp, _ := strconv.ParseInt(digs[0], 10, 32)
					mark.Quarter = int(mp)
				}
				pm, _ := strconv.ParseInt(fmark, 10, 32)
				mark.Grade = append(mark.Grade, int8(pm))
				marks = append(marks, mark)
			} else if sj.HasClass("itg-y") {
				_, fmark := data()
				mark.Annual = true
				pm, _ := strconv.ParseInt(fmark, 10, 32)
				mark.Grade = append(mark.Grade, int8(pm))
				marks = append(marks, mark)
			}
		})
	})

	return
}

// GetMessagesCount get current user messages count
func (cli *Client) GetMessagesCount() (unread int, total int, err error) {
	resp, err := cli.http.Get(fmt.Sprintf("%s/messages_count/", urlAjax))
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	respMap := make(map[string]int)
	json.Unmarshal([]byte(body), &respMap)

	return respMap["unread_messages"], respMap["all_messages"], nil
}

// GetMessages list for current user
func (cli *Client) GetMessages() (messages []Message, err error) {
	resp, err := cli.http.Get(urlMessages)
	if err != nil {
		return
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	pagesFlag := doc.Find("#content > div.pager > span.page_remark").Text()
	if pagesFlag != "" {
		pages := doc.Find("#content > div.pager > span.page")
		totalPages, ep := strconv.ParseInt(pages.Eq(pages.Size()-2).Text(), 10, 32)
		if ep != nil {
			err = ep
			return
		}
		if DEBUG {
			log.Printf("pages - '%s'\n", pages.Text())
			pages.Each(func(i int, p *goquery.Selection) {
				log.Printf("page - %s\n", p.Text())
			})
			log.Printf("total pages - %d\n", totalPages)
		}
	}

	doc.Find("#content > form > table.list > tbody > tr").Each(func(i int, s *goquery.Selection) {
		message := Message{}
		message.UserID = cli.Username
		msgID, _ := s.Find("td:nth-child(1) > input").Attr("value")
		message.ID, _ = strconv.ParseInt(msgID, 10, 64)
		title := s.Find("td:nth-child(2) > a")
		message.Subject = strings.TrimSpace(title.Text())
		if title.HasClass("unread") {
			message.IsUnread = true
		}
		from := s.Find("td:nth-child(3)").Text()
		message.From = from
		date := s.Find("td:nth-child(4)").Text()
		message.Date = russiantime.ParseDateString(date)
		messages = append(messages, message)
	})
	return
}

// GetMessage by id
func (cli *Client) GetMessage(msgID int64) (m Message, err error) {
	m.ID = msgID
	resp, err := cli.http.Get(fmt.Sprintf("%s/%d/", urlMessages, msgID))
	if err != nil {
		return
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	msgDate := strings.TrimPrefix(doc.Find("#msgview > div.msg-meta > div.msg-props > div:nth-child(1)").First().Text(), "Дата: ")
	m.Date = russiantime.ParseDateString(msgDate)
	msgFrom := doc.Find("#msgview > div.msg-meta > div.msg-props > div:nth-child(2) > a:nth-child(2)").First().Text()
	m.From = msgFrom
	msgText := doc.Find("#msgview > div.msg-text").First()
	m.Body = msgText.Text()

	return
}

func getClassName(s string) (classNumber int, classChar string, err error) {
	re, err := regexp.Compile(`\s?\n\s+Учащийся\n\s+\((\d+) "(.)"\)\n\s+`)
	if err != nil {
		return
	}
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		clsNum, _ := strconv.ParseInt(matches[1], 10, 32)
		classNumber = int(clsNum)
		classChar = re.FindStringSubmatch(s)[2]
	} else {
		return 0, "", errors.New("match size less or equal then 1")
	}

	return
}

// GetHomework to get user homework
func (cli *Client) GetHomework() (hws []Homework, err error) {
	resp, err := cli.http.Get(urlHomework)
	if err != nil {
		return
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	err = cli.getCurrentInfo()
	if err != nil {
		return
	}

	classIDText, _ := doc.Find("body").Attr("onload")
	if classIDText != "" {
		classIDText = strings.TrimRight(strings.TrimPrefix(classIDText, sLoadSubjectsS), sLoadSubjectsE)
	}
	classID, err := strconv.ParseInt(classIDText, 10, 64)
	hwPagesFlag := doc.Find("#homework_list > div.pager > span.page_remark").Text()
	if hwPagesFlag != "" {
		pages := doc.Find("#homework_list > div.pager > span.page")
		if DEBUG {
			log.Printf("total pages - %s\n", pages.Eq(pages.Size()-2).Text())
		}
	}

	doc.Find("#homework_list > table.list > tbody > tr").Each(func(i int, s *goquery.Selection) {
		h := Homework{}
		h.SchoolID = cli.SchoolID
		h.ClassID = classID
		date := s.Find("td:nth-child(1)").Text()
		h.Date = russiantime.ParseDateString(date)
		wday := s.Find("td:nth-child(2)").Text()
		h.DayOfWeek = wday
		course := s.Find("td:nth-child(3) > a").Text()
		h.CourseName = course
		hw := s.Find("td:nth-child(4)").Text()
		h.Homework = strings.TrimSpace(hw)
		subject := s.Find("td:nth-child(5)").Text()
		h.Subject = strings.TrimSpace(subject)

		hws = append(hws, h)
	})

	return
}

// GetTeachers to get class teachers
func (cli *Client) GetTeachers() (teachers []Teacher, err error) {
	resp, err := cli.http.Get(urlTeachers)
	if err != nil {
		return
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	doc.Find("#content > table.list > tbody > tr").Each(func(i int, t *goquery.Selection) {
		teacher := Teacher{}
		teacher.SchoolID = cli.SchoolID
		el := t.Find("td.action_links > a.mailto")
		href, _ := el.Attr("href")
		teacher.UserID = strings.TrimSuffix(
			strings.TrimPrefix(href, "/messages/new/?to="), fmt.Sprintf("@%d", cli.SchoolID))
		fio := t.Find("td:nth-child(2)").Text()
		teacher.FullName = fio
		courseName := strings.TrimSpace(t.Find("td:nth-child(3) > b").Text())
		if courseName == "" {
			courseName = strings.TrimSpace(t.Find("td:nth-child(3)").Text())
		}
		teacher.CourseName = courseName
		teachers = append(teachers, teacher)
	})

	return
}
