package endpoint

import (
	"fmt"
	"strings"

	"dhbw-loerrach.de/dualis/microservice/internal"
	"dhbw-loerrach.de/dualis/microservice/internal/api/models"
	"dhbw-loerrach.de/dualis/microservice/internal/api/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
)

func HandleStudentPerformance(params operations.StudentPerformanceParams, principal interface{}) middleware.Responder {

	var (
		enrollmentId     int64
		enrollmentGrade  string
		enrollmentStatus string
		lectureExamType  string
		lectureGrade     string
		lectureName      string
		lectureNumber    string
		lecturePresence  bool
		lectureWeighting string
		moduleName       string
		moduleNumber     string
		moduleCredits    string
		isWintersemester bool
		year             string
	)

	var errMsg *string
	lectureResults := []*models.LectureResult{}
	moduleResults := []*models.ModuleResult{}
	enrollments := []*models.Enrollment{}
	performancesList := models.PerformancesList{}

	claims := principal.(*internal.DualisClaims)
	filterableValues := map[string]interface{}{"student_fk": claims.StudentID}

	if params.IsWintersemester != nil {
		filterableValues["is_wintersemester"] = *params.IsWintersemester
	}

	if params.Year != nil {
		filterableValues["year"] = *params.Year
	}

	var values []interface{}
	var where []string
	for a, b := range filterableValues {
		where = append(where, fmt.Sprintf("%s = ?", a))
		var pVal *interface{}
		pVal = &b
		values = append(values, *pVal)
	}

	rows, err := db.Query("SELECT enrollment.id, grade, status, is_wintersemester, year, no, name, credits from dualis.enrollment INNER JOIN dualis.semester ON dualis.enrollment.semester_fk = dualis.semester.id INNER JOIN dualis.module ON dualis.enrollment.module_fk = dualis.module.id WHERE "+strings.Join(where, " AND "), values...)

	if err != nil {
		*errMsg = err.Error()
		return operations.NewStudentsInternalServerError().WithPayload(&models.SimpleError{Error: errMsg})
	}

	defer rows.Close()

	for rows.Next() {

		err = rows.Scan(&enrollmentId, &enrollmentGrade, &enrollmentStatus, &isWintersemester, &year, &moduleNumber, &moduleName, &moduleCredits)

		if err != nil {
			*errMsg = err.Error()
			return operations.NewStudentPerformanceInternalServerError().WithPayload(&models.SimpleError{Error: errMsg})
		}

		enrollmentIdVal := enrollmentId

		rows2, err := db.Query("SELECT exam_type, lecture_result.grade, name, no, presence, weighting FROM lecture_result INNER JOIN enrollment ON lecture_result.enrollment_fk = enrollment.id INNER JOIN lecture ON lecture_result.lecture_fk = lecture.id WHERE enrollment.id = ?", enrollmentIdVal)

		if err != nil {
			*errMsg = err.Error()
			return operations.NewStudentsInternalServerError().WithPayload(&models.SimpleError{Error: errMsg})
		}

		defer rows2.Close()

		for rows2.Next() {

			err = rows2.Scan(&lectureExamType, &lectureGrade, &lectureName, &lectureNumber, &lecturePresence, &lectureWeighting)

			if err != nil {
				*errMsg = err.Error()
				return operations.NewStudentPerformanceInternalServerError().WithPayload(&models.SimpleError{Error: errMsg})
			}

			lectureExamTypeVal := lectureExamType
			lectureGradeVal := lectureGrade
			lectureNameVal := lectureName
			lectureNumberVal := lectureNumber
			lecturePresenceVal := lecturePresence
			lectureWeightingVal := lectureWeighting

			lectureResults = append(lectureResults, &models.LectureResult{ExamType: &lectureExamTypeVal, Grade: lectureGradeVal, Name: &lectureNameVal, Number: &lectureNumberVal, Presence: lecturePresenceVal, Weighting: &lectureWeightingVal})
		}

		moduleCreditsVal := moduleCredits
		moduleNameVal := moduleName
		moduleNumberVal := moduleNumber

		moduleResults = append(moduleResults, &models.ModuleResult{Credits: &moduleCreditsVal, LectureResults: lectureResults, Name: &moduleNameVal, Number: &moduleNumberVal})
		lectureResults = nil

		enrollmentGradeVal := enrollmentGrade
		enrollmentStatusVal := enrollmentStatus
		isWintersemesterVal := isWintersemester
		yearVal := year

		var semester string

		if isWintersemesterVal {
			semester = "WiSe " + yearVal
		} else {
			semester = "SoSe " + yearVal
		}

		enrollments = append(enrollments, &models.Enrollment{Grade: enrollmentGradeVal, ModuleResult: moduleResults, Semester: &semester, Status: &enrollmentStatusVal})
		moduleResults = nil
	}

	performancesList = models.PerformancesList{Enrollments: enrollments}

	if len(enrollments) == 0 {
		return operations.NewStudentPerformanceNoContent()
	}

	return operations.NewStudentPerformanceOK().WithPayload(&performancesList)

}
