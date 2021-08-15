package httpsvc

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func (s *Server) handleCreateAssignment(c echo.Context) error {
	user := getUserFromCtx(c)
	assignmentReq := &assignmentReq{}
	err := c.Bind(assignmentReq)
	if err != nil {
		logrus.Error(err)
		return responseError(c, err)
	}

	assignment := assignmentCreateReqToModel(assignmentReq)
	assignment.AssignedBy = user.ID
	err = s.assignmentUsecase.Create(c.Request().Context(), assignment)
	if err != nil {
		logrus.Error(err)
		return responseError(c, err)
	}

	return c.JSON(http.StatusOK, assignmentModelToRes(assignment))
}

func (s *Server) handleDeleteAssignment(c echo.Context) error {
	id := c.Param("id")
	assignment, err := s.assignmentUsecase.DeleteByID(c.Request().Context(), id)
	if err != nil {
		logrus.Error(err)
		return responseError(c, err)
	}

	return c.JSON(http.StatusOK, assignmentModelToDeleteRes(assignment))
}

func (s *Server) handleGetAssignment(c echo.Context) error {
	id := c.Param("id")
	assignment, err := s.assignmentUsecase.FindByID(c.Request().Context(), id)
	if err != nil {
		logrus.Error(err)
		return responseError(c, err)
	}

	return c.JSON(http.StatusOK, assignmentModelToRes(assignment))
}

func (s *Server) handleGetAllAssignments(c echo.Context) error {
	cursor := getCursorFromContext(c)
	assignments, count, err := s.assignmentUsecase.FindAll(c.Request().Context(), cursor)
	if err != nil {
		logrus.Error(err)
		return responseError(c, err)
	}

	assignmentResponses := newAssignmentResponses(assignments)

	return c.JSON(http.StatusOK, newCursorRes(cursor, assignmentResponses, count))
}

func (s *Server) handleGetMyAssignments(c echo.Context) error {
	cursor := getCursorFromContext(c)
	assignments, count, err := s.assignmentUsecase.FindAll(c.Request().Context(), cursor)
	if err != nil {
		logrus.Error(err)
		return responseError(c, err)
	}

	res := newAssignmentResponses(assignments)
	for i := range res {
		// redacted
		res[i].CaseInputFileURL = ""
		res[i].CaseOutputFileURL = ""
	}

	return c.JSON(http.StatusOK, newCursorRes(cursor, res, count))
}

func (s *Server) handleGetAssignmentSubmissions(c echo.Context) error {
	id := c.Param("id")
	cursor := getCursorFromContext(c)
	submissions, count, err := s.assignmentUsecase.FindSubmissionsByID(c.Request().Context(), cursor, id)
	if err != nil {
		logrus.Error(err)
		return responseError(c, err)
	}

	submissionRes := newSubmissionResponses(submissions)

	return c.JSON(http.StatusOK, newCursorRes(cursor, submissionRes, count))
}

func (s *Server) handleUpdateAssignment(c echo.Context) error {
	assignmentReq := &assignmentReq{}
	err := c.Bind(assignmentReq)
	if err != nil {
		logrus.Error(err)
		return responseError(c, err)
	}

	assignment := assigmentUpdateReqToModel(assignmentReq)
	err = s.assignmentUsecase.Update(c.Request().Context(), assignment)
	if err != nil {
		logrus.Error(err)
		return responseError(c, err)
	}

	return c.JSON(http.StatusOK, assignmentModelToRes(assignment))

}
