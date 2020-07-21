/*
 * SPDX-License-Identifier: Apache-2.0
 */

package microfabd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/IBM-Blockchain/microfab/internal/pkg/blocks"
	"github.com/IBM-Blockchain/microfab/internal/pkg/channel"
	"github.com/IBM-Blockchain/microfab/internal/pkg/console"
	"github.com/IBM-Blockchain/microfab/internal/pkg/orderer"
	"github.com/IBM-Blockchain/microfab/internal/pkg/organization"
	"github.com/IBM-Blockchain/microfab/internal/pkg/peer"
	"github.com/IBM-Blockchain/microfab/internal/pkg/proxy"
	"github.com/hyperledger/fabric-protos-go/common"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

// Microfab represents an instance of the Microfab application.
type Microfab struct {
	sync.Mutex
	config                 *Config
	ordererOrganization    *organization.Organization
	endorsingOrganizations []*organization.Organization
	organizations          []*organization.Organization
	orderer                *orderer.Orderer
	peers                  []*peer.Peer
	genesisBlocks          map[string]*common.Block
	console                *console.Console
	proxy                  *proxy.Proxy
}

// New creates an instance of the Microfab application.
func New() (*Microfab, error) {
	config, err := DefaultConfig()
	if err != nil {
		return nil, err
	}
	fablet := &Microfab{
		config: config,
	}
	return fablet, nil
}

// Run runs the Fablet application.
func (m *Microfab) Run() error {

	// Grab the start time and say hello.
	startTime := time.Now()
	log.Print("Starting Microfab ...")

	// Ensure anything we start is stopped.
	defer m.stop()

	// Ensure the directory exists and is empty.
	err := m.ensureDirectory()
	if err != nil {
		return err
	}

	// Create all of the organizations.
	ctx := context.Background()
	eg, _ := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return m.createOrderingOrganization(m.config.OrderingOrganization)
	})
	for i := range m.config.EndorsingOrganizations {
		organization := m.config.EndorsingOrganizations[i]
		eg.Go(func() error {
			return m.createEndorsingOrganization(organization)
		})
	}
	err = eg.Wait()
	if err != nil {
		return err
	}

	// Sort the list of organizations by name, and then join all the organizations together.
	sort.Slice(m.endorsingOrganizations, func(i, j int) bool {
		return m.endorsingOrganizations[i].Name() < m.endorsingOrganizations[j].Name()
	})
	m.organizations = append(m.organizations, m.ordererOrganization)
	m.organizations = append(m.organizations, m.endorsingOrganizations...)

	// Create and start all of the components (orderer, peers).
	eg.Go(func() error {
		return m.createAndStartOrderer(m.ordererOrganization, 7050, 7051)
	})
	for i := range m.endorsingOrganizations {
		organization := m.endorsingOrganizations[i]
		apiPort := 7052 + (i * 3)
		chaincodePort := 7053 + (i * 3)
		operationsPort := 7054 + (i * 3)
		eg.Go(func() error {
			return m.createAndStartPeer(organization, apiPort, chaincodePort, operationsPort)
		})
	}
	err = eg.Wait()
	if err != nil {
		return err
	}

	// Sort the list of peers by their organization name.
	sort.Slice(m.peers, func(i, j int) bool {
		return m.peers[i].Organization().Name() < m.peers[j].Organization().Name()
	})

	// Create and start the console.
	console, err := console.New(m.organizations, m.orderer, m.peers, 8081, fmt.Sprintf("http://console.%s:%d", m.config.Domain, m.config.Port))
	if err != nil {
		return err
	}
	m.console = console
	go console.Start()

	// Create and start the proxy.
	proxy, err := proxy.New(console, m.orderer, m.peers, m.config.Port)
	if err != nil {
		return err
	}
	m.proxy = proxy
	go proxy.Start()

	// Connect to all of the components.
	channelCreator := m.endorsingOrganizations[0]
	err = m.orderer.Connect(channelCreator.MSP().ID(), channelCreator.Admin())
	if err != nil {
		return err
	}
	defer m.orderer.Close()
	for _, peer := range m.peers {
		err = peer.Connect(peer.Organization().MSP().ID(), peer.Organization().Admin())
		if err != nil {
			return err
		}
	}

	// Create and join all of the channels.
	for i := range m.config.Channels {
		channel := m.config.Channels[i]
		eg.Go(func() error {
			return m.createAndJoinChannel(channel)
		})
	}
	err = eg.Wait()
	if err != nil {
		return err
	}

	// Say how long start up took, then wait for signals.
	readyTime := time.Now()
	startupDuration := readyTime.Sub(startTime)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Microfab started in %vms", startupDuration.Milliseconds())
	<-sigs
	log.Printf("Stopping Microfab due to signal ...")
	m.stop()
	log.Printf("Microfab stopped")
	return nil

}

func (m *Microfab) ensureDirectory() error {
	if m.directoryExists() {
		err := m.removeDirectory()
		if err != nil {
			return err
		}
	} else {
		err := m.createDirectory()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Microfab) directoryExists() bool {
	if _, err := os.Stat(m.config.Directory); os.IsNotExist(err) {
		return false
	}
	return true
}

func (m *Microfab) createDirectory() error {
	return os.MkdirAll(m.config.Directory, 0755)
}

func (m *Microfab) removeDirectory() error {
	file, err := os.Open(m.config.Directory)
	if err != nil {
		return err
	}
	defer file.Close()
	names, err := file.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(path.Join(m.config.Directory, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Microfab) createOrderingOrganization(config Organization) error {
	log.Printf("Creating ordering organization %s ...", config.Name)
	organization, err := organization.New(config.Name)
	if err != nil {
		return err
	}
	m.Lock()
	defer m.Unlock()
	m.ordererOrganization = organization
	log.Printf("Created ordering organization %s", config.Name)
	return nil
}

func (m *Microfab) createEndorsingOrganization(config Organization) error {
	log.Printf("Creating endorsing organization %s ...", config.Name)
	organization, err := organization.New(config.Name)
	if err != nil {
		return err
	}
	m.Lock()
	defer m.Unlock()
	m.endorsingOrganizations = append(m.endorsingOrganizations, organization)
	log.Printf("Created endorsing organization %s", config.Name)
	return nil
}

func (m *Microfab) createAndStartOrderer(organization *organization.Organization, apiPort, operationsPort int) error {
	log.Printf("Creating and starting orderer for ordering organization %s ...", organization.Name())
	directory := path.Join(m.config.Directory, "orderer")
	orderer, err := orderer.New(
		organization,
		directory,
		apiPort,
		fmt.Sprintf("grpc://orderer-api.%s:%d", m.config.Domain, m.config.Port),
		operationsPort,
		fmt.Sprintf("http://orderer-operations.%s:%d", m.config.Domain, m.config.Port),
	)
	if err != nil {
		return err
	}
	err = orderer.Start(m.endorsingOrganizations)
	if err != nil {
		return err
	}
	m.Lock()
	defer m.Unlock()
	m.orderer = orderer
	log.Printf("Created and started orderer for ordering organization %s", organization.Name())
	return nil
}

func (m *Microfab) createAndStartPeer(organization *organization.Organization, apiPort, chaincodePort, operationsPort int) error {
	log.Printf("Creating and starting peer for ordering organization %s ...", organization.Name())
	organizationName := organization.Name()
	lowerOrganizationName := strings.ToLower(organizationName)
	peerDirectory := path.Join(m.config.Directory, fmt.Sprintf("peer-%s", lowerOrganizationName))
	peer, err := peer.New(
		organization,
		peerDirectory,
		int32(apiPort),
		fmt.Sprintf("grpc://%speer-api.%s:%d", lowerOrganizationName, m.config.Domain, m.config.Port),
		int32(chaincodePort),
		fmt.Sprintf("grpc://%speer-chaincode.%s:%d", lowerOrganizationName, m.config.Domain, m.config.Port),
		int32(operationsPort),
		fmt.Sprintf("http://%speer-operations.%s:%d", lowerOrganizationName, m.config.Domain, m.config.Port),
	)
	if err != nil {
		return err
	}
	err = peer.Start()
	if err != nil {
		return err
	}
	m.Lock()
	defer m.Unlock()
	m.peers = append(m.peers, peer)
	log.Printf("Created and started peer for endorsing organization %s", organization.Name())
	return nil
}

func (m *Microfab) createChannel(config Channel) (*common.Block, error) {
	log.Printf("Creating channel %s ...", config.Name)
	opts := []channel.Option{
		channel.WithCapabilityLevel(m.config.CapabilityLevel),
	}
	for _, endorsingOrganization := range m.endorsingOrganizations {
		found := false
		for _, organizationName := range config.EndorsingOrganizations {
			if endorsingOrganization.Name() == organizationName {
				found = true
				break
			}
		}
		if found {
			opts = append(opts, channel.AddMSPID(endorsingOrganization.MSP().ID()))
		}
	}
	err := channel.CreateChannel(m.orderer, config.Name, opts...)
	if err != nil {
		return nil, err
	}
	var genesisBlock *common.Block
	for {
		genesisBlock, err = blocks.GetGenesisBlock(m.orderer, config.Name)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}
	opts = []channel.Option{}
	for _, peer := range m.peers {
		found := false
		for _, organizationName := range config.EndorsingOrganizations {
			if peer.Organization().Name() == organizationName {
				found = true
				break
			}
		}
		if found {
			opts = append(opts, channel.AddAnchorPeer(peer.MSPID(), peer.Hostname(), peer.Port()))
		}
	}
	err = channel.UpdateChannel(m.orderer, config.Name, opts...)
	if err != nil {
		return nil, err
	}
	log.Printf("Created channel %s", config.Name)
	return genesisBlock, nil
}

func (m *Microfab) createAndJoinChannel(config Channel) error {
	log.Printf("Creating and joining channel %s ...", config.Name)
	genesisBlock, err := m.createChannel(config)
	if err != nil {
		return err
	}
	ctx := context.Background()
	eg, _ := errgroup.WithContext(ctx)
	for i := range m.peers {
		peer := m.peers[i]
		found := false
		for _, organizationName := range config.EndorsingOrganizations {
			if peer.Organization().Name() == organizationName {
				found = true
				break
			}
		}
		if found {
			eg.Go(func() error {
				log.Printf("Joining channel %s on peer for endorsing organization %s ...", config.Name, peer.Organization().Name())
				err := peer.JoinChannel(genesisBlock)
				if err != nil {
					return err
				}
				log.Printf("Joined channel %s on peer for endorsing organization %s", config.Name, peer.Organization().Name())
				return nil
			})
		}
	}
	err = eg.Wait()
	if err != nil {
		return err
	}
	log.Printf("Created and joined channel %s", config.Name)
	return nil
}

func (m *Microfab) stop() error {
	if m.proxy != nil {
		err := m.proxy.Stop()
		if err != nil {
			return err
		}
		m.proxy = nil
	}
	if m.console != nil {
		err := m.console.Stop()
		if err != nil {
			return err
		}
		m.console = nil
	}
	for _, peer := range m.peers {
		err := peer.Stop()
		if err != nil {
			return err
		}
	}
	m.peers = []*peer.Peer{}
	if m.orderer != nil {
		err := m.orderer.Stop()
		if err != nil {
			return err
		}
		m.orderer = nil
	}
	return nil
}
