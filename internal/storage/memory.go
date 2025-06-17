package storage

import (
	"sync"

	pb "github.com/ankittk/email-audit-service/proto"
)

var (
	auditReports = make(map[string]*pb.AuditReport)
	mutex        = sync.RWMutex{}
)

func StoreAuditReport(report *pb.AuditReport) {
	mutex.Lock()
	defer mutex.Unlock()
	auditReports[report.AuditId] = report
}

func GetAuditReport(auditID string) *pb.AuditReport {
	mutex.RLock()
	defer mutex.RUnlock()
	return auditReports[auditID]
}

func GetAuditSummary(userID, companyID string, page, pageSize int) ([]*pb.AuditReport, int) {
	mutex.RLock()
	defer mutex.RUnlock()

	var filtered []*pb.AuditReport
	for _, report := range auditReports {
		if (userID == "" || report.UserId == userID) &&
			(companyID == "" || report.CompanyId == companyID) {
			filtered = append(filtered, report)
		}
	}

	total := len(filtered)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*pb.AuditReport{}, total
	}

	if end > total {
		end = total
	}

	return filtered[start:end], total
}
