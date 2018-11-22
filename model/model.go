package model

import (
	"context"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/client"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const PathSep = "/"

/* Model definition and initialization */

type model struct{
	etcdEndpoints []string
	etcdTO int
	etcdRootNs string
}

func NewModel(etcdEndpoints []string, etcdTO int, etcdRootNs string) *model {
	return &model{
		etcdEndpoints: etcdEndpoints,
		etcdTO: etcdTO,
		etcdRootNs: etcdRootNs,
	}
}

/* Model data structures */

type Mutex struct {
	Name string
	Hostname string
	Timestamp time.Time
	Description string
	EtcdPath string
}

type Service struct {
	Name string
	Mutexes []*Mutex
}

/* etcd cluster connection handling */

func (m *model) initEtcdClient(endpoints []string) (client.KeysAPI, error) {
	cfg := client.Config {
		Endpoints: endpoints,
	}
	c, err := client.New(cfg)
	if err != nil {
		return nil, err
	}
	cli := client.NewKeysAPI(c)
	return cli, nil
}

func (m *model) getTOContext() (context.Context, context.CancelFunc) {
	to := time.Duration(m.etcdTO) * time.Second
	return context.WithTimeout(context.Background(), to)
}

/* etcd tree hierarchy manipulation */

/* Get the last component of the etcd key path */
func pathLast(path string) string {
	return path[strings.LastIndex(path, PathSep) + 1:]
}

/* Get descendant of given root node by its relative path from root */
func getDesc(root *client.Node, path string) *client.Node {
	if !root.Dir {
		return nil
	}
	i := strings.Index(path, PathSep)
	if i == -1 {
		for _, node := range root.Nodes {
			if pathLast(node.Key) == path {
				return node
			}
		}
	} else {
		for _, node := range root.Nodes {
			if pathLast(node.Key) == path[:i] {
				return getDesc(node, path[i + 1:])
			}
		}
	}
	return nil
}

/* Get the descendant's value as string */
func getDescStrVal(root *client.Node, path string) string {
	node := getDesc(root, path)
	if node == nil {
		return ""
	} else {
		return node.Value
	}
}

/* Get the descendant's value as time */
func getDescTimeVal(root *client.Node, path string) time.Time {
	timeStr := getDescStrVal(root, path)
	if timeStr == "" {
		return time.Time{}
	}
	tStamp, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return time.Time{}
	} else {
		return time.Unix(tStamp, 0)
	}
}

/* Model public API */

/* Test all endpoints by issuing `get` request with root namespace key */
func (m *model) TestConnection() error {
	for _, endpoint := range m.etcdEndpoints {
		cli, err := m.initEtcdClient([]string{endpoint})
		if err != nil {
			return err
		}
		ctx, cancel := m.getTOContext()
		defer cancel()
		_, err = cli.Get(ctx, m.etcdRootNs, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

/* List all mutexes from the root namespace and group them by services */
func (m *model) ListMutexes() ([]*Service, error) {
	cli, err := m.initEtcdClient(m.etcdEndpoints)
	if err != nil {
		return nil, err
	}
	ctx, cancel := m.getTOContext()
	defer cancel()
	ns := m.etcdRootNs
	resp, err := cli.Get(ctx, ns, &client.GetOptions{Recursive: true, Sort: true})
	if err != nil {
		if client.IsKeyNotFound(err) {
			return nil, errors.Wrapf(err, "could not open etcd root namespace '%s'", ns)
		} else {
			return nil, err
		}
	}
	services := make([]*Service, resp.Node.Nodes.Len())
	for i, serviceNode := range resp.Node.Nodes {
		if serviceNode.Dir {
			mutexes := make([]*Mutex, 0)
			for _, mutexNode := range serviceNode.Nodes {
				if mutexNode.Dir && getDesc(mutexNode, "held") != nil {
					mutexes = append(mutexes, &Mutex{
						Name: pathLast(mutexNode.Key),
						Hostname: getDescStrVal(mutexNode, "owner/hostname"),
						Timestamp: getDescTimeVal(mutexNode, "owner/since"),
						Description: getDescStrVal(mutexNode, "owner/help"),
						EtcdPath: mutexNode.Key,
					})
				}
			}
			services[i] = &Service{
				Name: pathLast(serviceNode.Key),
				Mutexes: mutexes,
			}
		}
	}
	return services, nil
}

/* Delete (unlock) mutex from etcd cluster and all of its metadata */
func (m *model) UnlockMutex(mutexPath string) error {
	cli, err := m.initEtcdClient(m.etcdEndpoints)
	if err != nil {
		return err
	}
	ctx, cancel := m.getTOContext()
	defer cancel()
	_, err = cli.Delete(ctx, mutexPath, &client.DeleteOptions{ Recursive: true })
	if err != nil && client.IsKeyNotFound(err) {
		return errors.Wrapf(err, "mutex named '%s' does not exist", mutexPath)
	}
	return err
}



/* Dummy queries */

func (m *model) DummyMutexList() []*Mutex {
	mutexes := make([]*Mutex, 2)
	mutexes[0] = &Mutex{ Name: "Delete", Hostname: "SRV-1234", Timestamp: time.Unix(int64(rand.Int31()), 0), Description: "Test", EtcdPath: "/mutexes/user_profile/logout/held" }
	mutexes[1] = &Mutex{ Name: "Update", Hostname: "SRV-5678", Timestamp: time.Unix(int64(rand.Int31()), 0), Description: "Hello world", EtcdPath: "/mutexes/billing/pay_cash/held" }
	return mutexes
}

func (m *model) DummyServiceList() []*Service {
	services := make([]*Service, 2)
	services[0] = &Service{ Name: "Servica 1", Mutexes: m.DummyMutexList() }
	services[1] = &Service{ Name: "Servica 2", Mutexes: m.DummyMutexList() }
	return services
}

