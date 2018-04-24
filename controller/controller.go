package controller

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/contrib/pkg/broker/controller"
	"github.com/kubernetes-incubator/service-catalog/contrib/pkg/brokerapi"
	"github.com/philhug/mariadb-broker/client"
)

type Config struct {
	DBUser string
	DBPass string
}

type errNoSuchInstance struct {
	instanceID string
}

func (e errNoSuchInstance) Error() string {
	return fmt.Sprintf("no such instance with ID %s", e.instanceID)
}

type mariadbController struct {
	config Config
	client client.Client
}

// CreateController creates an instance of a User Provided service broker controller.
func CreateController(config Config) controller.Controller {
	dsn := fmt.Sprintf("root:%s@tcp(mysql:3306)/", config.DBPass)
	fmt.Println(dsn)
	return &mariadbController{config: config, client: client.NewClient(client.Config{Dsn: dsn})}
}

func (c *mariadbController) Catalog() (*brokerapi.Catalog, error) {
	return &brokerapi.Catalog{
		Services: []*brokerapi.Service{
			{
				Name:        "mariadb",
				ID:          "3533e2f0-6335-4a4e-9d15-d7c0b90b75b5",
				Description: "MariaDB database",
				Plans: []brokerapi.ServicePlan{
					{
						Name:        "default",
						ID:          "b9600ecb-d511-4621-b450-a0fa1738e632",
						Description: "MariaDB database",
						Free:        true,
					},
				},
				Bindable: true,
			},
		},
	}, nil
}

func (c *mariadbController) CreateServiceInstance(instanceID string, req *brokerapi.CreateServiceInstanceRequest) (*brokerapi.CreateServiceInstanceResponse, error) {
	glog.Infof("CreateServiceInstance")
	return &brokerapi.CreateServiceInstanceResponse{}, nil
}

func (c *mariadbController) GetServiceInstance(id string) (string, error) {
	return "", errors.New("Unimplemented")
}

func (c *mariadbController) RemoveServiceInstance(instanceID, serviceID, planID string, acceptsIncomplete bool) (*brokerapi.DeleteServiceInstanceResponse, error) {
/*
	if err := client.Delete(releaseName(instanceID)); err != nil {
		return nil, err
	}
*/
	return &brokerapi.DeleteServiceInstanceResponse{}, nil
}

func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func (c *mariadbController) Bind(instanceID, bindingID string, req *brokerapi.BindingRequest) (*brokerapi.CreateServiceBindingResponse, error) {
	host := releaseName(instanceID) + "-mariadb." + instanceID + ".svc.cluster.local"
	port := "3306"
	database := releaseName(bindingID)
	username := string(bindingID[0:31])
/*
	password, err := client.GetPassword(releaseName(bindingID), instanceID)
	if err != nil {
		return nil, err
	}
*/
	if err := c.client.CreateDB(database); err != nil {
		return nil, err
	}
	password := RandomString(32)
	if err := c.client.CreateUser(database, username, password); err != nil {
		return nil, err
	}
	return &brokerapi.CreateServiceBindingResponse{
		Credentials: brokerapi.Credential{
			"uri":      "mysql://" + username + ":" + password + "@" + host + ":" + port + "/" + database,
			"username": username,
			"password": password,
			"host":     host,
			"port":     port,
			"database_name": database,
		},
	}, nil
}

func (c *mariadbController) UnBind(instanceID, bindingID, serviceID, planID string) error {
	// Since we don't persist the binding, there's nothing to do here.
	return nil
}

func (c *mariadbController) GetServiceInstanceLastOperation(instanceID, serviceID, planID, operation string) (*brokerapi.LastOperationResponse, error) {
	// TODO
	return nil, nil
}

func releaseName(id string) string {
	return "i-" + id
}
