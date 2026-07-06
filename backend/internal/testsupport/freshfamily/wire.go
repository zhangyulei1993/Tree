package freshfamily

import (
	"gorm.io/gorm"

	quotaservice "tree/backend/internal/accountquota/service"
	"tree/backend/internal/common/contentsafety"
	inviterepo "tree/backend/internal/family/invitation/repository"
	invitationservice "tree/backend/internal/family/invitation/service"
	joinrepo "tree/backend/internal/family/joinrequest/repository"
	joinrequestservice "tree/backend/internal/family/joinrequest/service"
	"tree/backend/internal/testsupport/freshquota"
)

func NewInvitationService(tx *gorm.DB, quota quotaservice.Service) invitationservice.Service {
	repo := inviterepo.NewRepository(tx)
	return invitationservice.NewService(repo, inviterepo.NewUnitOfWork(tx, repo), freshquota.AllowAllFamilyPerm{}, quota, contentsafety.AlwaysPass())
}

func NewJoinRequestService(tx *gorm.DB, quota quotaservice.Service) joinrequestservice.Service {
	repo := joinrepo.NewRepository(tx)
	return joinrequestservice.NewService(repo, joinrepo.NewUnitOfWork(tx, repo), freshquota.AllowAllFamilyPerm{}, quota, contentsafety.AlwaysPass())
}
