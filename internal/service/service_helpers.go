package service

import "e-memo-job-reservation-api/internal/model"

func determineUserContexts(user *model.Employee, ticket *model.Ticket, requestor *model.Employee, job *model.Job) []string {
	var contexts []string
	if user.NPK == ticket.Requestor {
		contexts = append(contexts, "SELF")
	}
	if user.DepartmentID == requestor.DepartmentID {
		contexts = append(contexts, "REQUESTOR_DEPT")
	}
	if user.DepartmentID == ticket.DepartmentTargetID {
		contexts = append(contexts, "TARGET_DEPT")
	}
	if job != nil && job.PicJob.Valid && user.NPK == job.PicJob.String {
		contexts = append(contexts, "ASSIGNED")
	}
	return contexts
}
