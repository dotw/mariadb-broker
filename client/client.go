package client

import (
	yaml "gopkg.in/yaml.v2"

	"database/sql"
	"github.com/golang/glog"
	"github.com/dchest/uniuri"
	_ "github.com/go-sql-driver/mysql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/helm"
)

const (
	tillerHost = "tiller-deploy.kube-system.svc.cluster.local:44134"
	chartPath  = "/mariadb-0.6.1.tgz"
)

type Config struct {
	// e.g. "root:root@tcp(mysql:3306)/"
	Dsn string
}

type Client struct {
	config Config
}

func NewClient(config Config) (Client) {
	return Client{config: config}
}

func (c *Client) CreateDB(database string) error {
	db, err := sql.Open("mysql", c.config.Dsn)
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	// TODO fix SQL injection
	glog.Infof("CREATE DATABASE `" + database + "`")
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS `" + database + "`")
	if err != nil {
		panic(err)
	}
	return nil
}

func (c *Client) CreateUser(database, username, password string) error {
	db, err := sql.Open("mysql", c.config.Dsn)
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	glog.Infof("GRANT USER")
	_, err = db.Exec("GRANT ALL PRIVILEGES ON `" + database + "`.* TO '"+ username + "'@'%' IDENTIFIED by '" + password + "'")
	if err != nil {
		panic(err)
	}

	return nil
}

// Install creates a new MariaDB chart release
func (c *Client) Install(releaseName, namespace string) error {
	vals, err := yaml.Marshal(map[string]interface{}{
		"mariadbRootPassword": uniuri.New(),
		"mariadbDatabase":     "dbname",
	})
	if err != nil {
		return err
	}
	helmClient := helm.NewClient(helm.Host(tillerHost))
	_, err = helmClient.InstallRelease(chartPath, namespace, helm.ReleaseName(releaseName), helm.ValueOverrides(vals))
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a MariaDB chart release
func (c *Client) Delete(releaseName string) error {
	helmClient := helm.NewClient(helm.Host(tillerHost))
	if _, err := helmClient.DeleteRelease(releaseName); err != nil {
		return err
	}
	return nil
}

// GetPassword returns the MariaDB password for a chart release
func (c *Client) GetPassword(releaseName, namespace string) (string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}
	secret, err := clientset.Core().Secrets(namespace).Get(releaseName+"-mariadb", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(secret.Data["mariadb-root-password"]), nil
}
