package usecase

import (
	"github.com/google/uuid"
	"sample-go-server/entity"
)

type ContractUseCase struct {
	repo entity.IContractRepository
}

func NewContractUseCase(r entity.IContractRepository) entity.IContractUseCase {
	return &ContractUseCase{repo: r}
}

func (uc ContractUseCase) CreateNew(c entity.NewContractRequest) (*entity.Contract, error) {

	contract := entity.Contract{
		Uuid:        uuid.New().String(),
		Name:        c.Name,
		UserUuid:    c.UserUuid,
		Description: c.Description,
	}

	return uc.repo.Add(contract)
}

func (uc ContractUseCase) GetListByUser(userUuid string) ([]entity.Contract, error) {
	return uc.repo.GetListByUser(userUuid)
}

func (uc ContractUseCase) Update(cUuid string, c entity.UpdateContractRequest) error {
	return uc.repo.Update(cUuid, c.ToMap())
}

func (uc ContractUseCase) GetDetails(cUuid string) (*entity.Contract, error) {
	return uc.repo.GetDetails(cUuid)
}

func (uc ContractUseCase) Delete(cUuid string) error {
	return uc.repo.Delete(cUuid)
}

func (uc ContractUseCase) DeleteByUser(cUuid string, uUuid string) error {
	return uc.repo.DeleteByUser(cUuid, uUuid)
}
