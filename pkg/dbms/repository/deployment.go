package repository

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
)

//DeploymentDAOInterface DeploymentDAOInterface
type DeploymentDAOInterface interface {
	CreateDeployment(deployment model.Deployment) (int, error)
	EditDeployment(deployment model.Deployment) error
	GetDeploymentByID(id int) (model.Deployment, error)
	ListDeployments(environmentID, requestDeploymentID string, pageNumber, pageSize int) ([]model.Deployments, error)
	CountDeployments(environmentID, requestDeploymentID string) (int64, error)
}

//DeploymentDAOImpl DeploymentDAOImpl
type DeploymentDAOImpl struct {
	Db *gorm.DB
}

//GetDeploymentByID GetDeploymentByID
func (dao DeploymentDAOImpl) GetDeploymentByID(id int) (model.Deployment, error) {
	var deployment model.Deployment
	if err := dao.Db.First(&deployment, id).Error; err != nil {
		return model.Deployment{}, err
	}
	return deployment, nil
}

//CreateDeployment create deployment
func (dao DeploymentDAOImpl) CreateDeployment(deployment model.Deployment) (int, error) {
	if err := dao.Db.Create(&deployment).Error; err != nil {
		return -1, err
	}
	return int(deployment.ID), nil
}

//EditDeployment edit deployment
func (dao DeploymentDAOImpl) EditDeployment(deployment model.Deployment) error {
	gorm := dao.Db.Save(&deployment)
	return gorm.Error
}

func prepareSQL(environmentID string) string {
	sql := "deployments.request_deployment_id = ?"
	if environmentID != "" {
		sql += " AND deployments.environment_id = " + environmentID
	}
	return sql
}

//ListDeployments list all deployments filtered by date, environment and user
func (dao DeploymentDAOImpl) ListDeployments(environmentID, requestDeploymentID string, pageNumber, pageSize int) ([]model.Deployments, error) {
	var deployments []model.Deployments
	sql := prepareSQL(environmentID)
	rows, err := dao.Db.Table("deployments").Select(
		"deployments.id AS id, deployments.created_at AS created_at, deployments.updated_at AS updated_at,chart, request_deployment_id, environments.id AS environments_id, environments.name AS environments_name, processed ,success, message ",
	).Joins(
		"JOIN environments ON deployments.environment_id = environments.id",
	).Where(sql, requestDeploymentID).Offset((pageNumber - 1) * pageSize).Limit(pageSize).Rows()

	for rows.Next() {
		reqID, envID, id := 0, 0, 0
		createdAt := time.Time{}
		updatedAt := time.Time{}
		chart, envName, message := "", "", ""
		success, processed := false, false
		rows.Scan(&id, &createdAt, &updatedAt, &chart, &reqID, &envID, &envName, &processed, &success, &message)

		deployment := model.Deployments{}
		deployment.ID = uint(id)
		deployment.CreatedAt = createdAt
		deployment.UpdatedAt = updatedAt
		deployment.Chart = chart
		deployment.Environment.ID = uint(envID)
		deployment.Environment.Name = envName
		deployment.Processed = processed
		deployment.RequestDeploymentID = uint(reqID)
		deployment.Success = success
		deployment.Message = message

		deployments = append(deployments, deployment)
	}

	return deployments, err
}

//CountDeployments count
func (dao DeploymentDAOImpl) CountDeployments(environmentID, requestDeploymentID string) (int64, error) {
	var deployment model.Deployment
	var count int64
	sql := prepareSQL(environmentID)
	err := dao.Db.Where(sql, requestDeploymentID).Model(&deployment).Count(&count).Error
	return count, err
}
