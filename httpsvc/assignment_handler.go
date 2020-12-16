package httpsvc

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/miun173/autograd/utils"
	"github.com/sirupsen/logrus"
)

func (s *Server) handleCreateAssignment(c echo.Context) error {
	assignmentReq := &assignmentReq{}
	err := c.Bind(assignmentReq)
	if err != nil {
		logrus.Error(err)
		return responseError(c, err)
	}

	assignment := assignmentCreateReqToModel(assignmentReq)
	err = s.assignmentUsecase.Create(c.Request().Context(), assignment)
	if err != nil {
		logrus.Error(err)
		return responseError(c, err)
	}

	return c.JSON(http.StatusOK, assignmentModelToCreateRes(assignment))
}

func (s *Server) handleDeleteAssignment(c echo.Context) error {
	id := utils.StringToInt64(c.Param("ID"))
	assignment, err := s.assignmentUsecase.Delete(c.Request().Context(), id)
	if err != nil {
		logrus.Error(err)
		return responseError(c, err)
	}

	return c.JSON(http.StatusOK, assignmentModelToDeleteRes(assignment))
}

func (s *Server) handleGetAssignment(c echo.Context) error {
	id := utils.StringToInt64(c.Param("ID"))
	assignment, err := s.assignmentUsecase.FindByID(c.Request().Context(), id)
	if err != nil {
		logrus.Error(err)
		return responseError(c, err)
	}

	return c.JSON(http.StatusOK, assignmentModelToCreateRes(assignment))
}

func (s *Server) handleGetAssignments(c echo.Context) error {
	cursor := getCursorFromContext(c)
	assignments, count, err := s.assignmentUsecase.FindAll(c.Request().Context(), cursor)
	if err != nil {
		logrus.Error(err)
		return responseError(c, err)
	}

	assignmentResponses := newAssignmentResponses(assignments)

	return c.JSON(http.StatusOK, newCursorRes(cursor, assignmentResponses, count))
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

	return c.JSON(http.StatusOK, assignmentModelToCreateRes(assignment))

}