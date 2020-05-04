/*
 * SPDX-License-Identifier: Apache-2.0
 */

package fablet

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

	"github.com/IBM-Blockchain/fablet/internal/pkg/blocks"
	"github.com/IBM-Blockchain/fablet/internal/pkg/channel"
	"github.com/IBM-Blockchain/fablet/internal/pkg/console"
	"github.com/IBM-Blockchain/fablet/internal/pkg/orderer"
	"github.com/IBM-Blockchain/fablet/internal/pkg/organization"
	"github.com/IBM-Blockchain/fablet/internal/pkg/peer"
	"github.com/IBM-Blockchain/fablet/internal/pkg/proxy"
	"github.com/hyperledger/fabric-protos-go/common"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

// Fablet represents an instance of the Fablet application.
type Fablet struct {
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

// New creates an instance of the Fablet application.
func New() (*Fablet, error) {
	config, err := DefaultConfig()
	if err != nil {
		return nil, err
	}
	fablet := &Fablet{
		config: config,
	}
	return fablet, nil
}

// Run runs the Fablet application.
func (f *Fablet) Run() error {

	// Grab the start time and say hello.
	startTime := time.Now()
	log.Print("Starting Fablet ...")

	// Ensure anything we start is stopped.
	defer f.stop()

	// Ensure the directory exists and is empty.
	err := f.ensureDirectory()
	if err != nil {
		return err
	}

	// Create all of the organizations.
	ctx := context.Background()
	eg, _ := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return f.createOrderingOrganization(f.config.OrderingOrganization)
	})
	for i := range f.config.EndorsingOrganizations {
		organization := f.config.EndorsingOrganizations[i]
		eg.Go(func() error {
			return f.createEndorsingOrganization(organization)
		})
	}
	err = eg.Wait()
	if err != nil {
		return err
	}

	// Sort the list of organizations by name, and then join all the organizations together.
	sort.Slice(f.endorsingOrganizations, func(i, j int) bool {
		return f.endorsingOrganizations[i].Name() < f.endorsingOrganizations[j].Name()
	})
	f.organizations = append(f.organizations, f.ordererOrganization)
	f.organizations = append(f.organizations, f.endorsingOrganizations...)

	// Create and start all of the components (orderer, peers).
	eg.Go(func() error {
		return f.createAndStartOrderer(f.ordererOrganization, 7050, 7051)
	})
	for i := range f.endorsingOrganizations {
		organization := f.endorsingOrganizations[i]
		apiPort := 7052 + (i * 3)
		chaincodePort := 7053 + (i * 3)
		operationsPort := 7054 + (i * 3)
		eg.Go(func() error {
			return f.createAndStartPeer(organization, apiPort, chaincodePort, operationsPort)
		})
	}
	err = eg.Wait()
	if err != nil {
		return err
	}

	// Sort the list of peers by their organization name.
	sort.Slice(f.peers, func(i, j int) bool {
		return f.peers[i].Organization().Name() < f.peers[j].Organization().Name()
	})

	// Create and start the console.
	console, err := console.New(f.organizations, f.orderer, f.peers, 8081, fmt.Sprintf("http://console.%s:%d", f.config.Domain, f.config.Port))
	if err != nil {
		return err
	}
	f.console = console
	go console.Start()

	// Create and start the proxy.
	proxy, err := proxy.New(console, f.orderer, f.peers, f.config.Port)
	if err != nil {
		return err
	}
	f.proxy = proxy
	go proxy.Start()

	// Connect to all of the components.
	channelCreator := f.endorsingOrganizations[0]
	err = f.orderer.Connect(channelCreator.MSP().ID(), channelCreator.Admin())
	if err != nil {
		return err
	}
	defer f.orderer.Close()
	for _, peer := range f.peers {
		err = peer.Connect(peer.Organization().MSP().ID(), peer.Organization().Admin())
		if err != nil {
			return err
		}
	}

	// Create and join all of the channels.
	for i := range f.config.Channels {
		channel := f.config.Channels[i]
		eg.Go(func() error {
			return f.createAndJoinChannel(channel)
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
	log.Printf("Fablet started in %vms", startupDuration.Milliseconds())
	<-sigs
	log.Printf("Stopping Fablet due to signal ...")
	f.stop()
	log.Printf("Fablet stopped")
	return nil

}

func (f *Fablet) ensureDirectory() error {
	if f.directoryExists() {
		err := f.removeDirectory()
		if err != nil {
			return err
		}
	} else {
		err := f.createDirectory()
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Fablet) directoryExists() bool {
	if _, err := os.Stat(f.config.Directory); os.IsNotExist(err) {
		return false
	}
	return true
}

func (f *Fablet) createDirectory() error {
	return os.MkdirAll(f.config.Directory, 0755)
}

func (f *Fablet) removeDirectory() error {
	file, err := os.Open(f.config.Directory)
	if err != nil {
		return err
	}
	defer file.Close()
	names, err := file.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(path.Join(f.config.Directory, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Fablet) createOrderingOrganization(config Organization) error {
	log.Printf("Creating ordering organization %s ...", config.Name)
	organization, err := organization.New(config.Name)
	if err != nil {
		return err
	}
	f.Lock()
	defer f.Unlock()
	f.ordererOrganization = organization
	log.Printf("Created ordering organization %s", config.Name)
	return nil
}

func (f *Fablet) createEndorsingOrganization(config Organization) error {
	log.Printf("Creating endorsing organization %s ...", config.Name)
	organization, err := organization.New(config.Name)
	if err != nil {
		return err
	}
	f.Lock()
	defer f.Unlock()
	f.endorsingOrganizations = append(f.endorsingOrganizations, organization)
	log.Printf("Created endorsing organization %s", config.Name)
	return nil
}

func (f *Fablet) createAndStartOrderer(organization *organization.Organization, apiPort, operationsPort int) error {
	log.Printf("Creating and starting orderer for ordering organization %s ...", organization.Name())
	directory := path.Join(f.config.Directory, "orderer")
	orderer, err := orderer.New(
		organization,
		directory,
		apiPort,
		fmt.Sprintf("grpc://orderer-api.%s:%d", f.config.Domain, f.config.Port),
		operationsPort,
		fmt.Sprintf("http://orderer-operations.%s:%d", f.config.Domain, f.config.Port),
	)
	if err != nil {
		return err
	}
	err = orderer.Start(f.endorsingOrganizations)
	if err != nil {
		return err
	}
	f.Lock()
	defer f.Unlock()
	f.orderer = orderer
	log.Printf("Created and started orderer for ordering organization %s", organization.Name())
	return nil
}

func (f *Fablet) createAndStartPeer(organization *organization.Organization, apiPort, chaincodePort, operationsPort int) error {
	log.Printf("Creating and starting peer for ordering organization %s ...", organization.Name())
	organizationName := organization.Name()
	lowerOrganizationName := strings.ToLower(organizationName)
	peerDirectory := path.Join(f.config.Directory, fmt.Sprintf("peer-%s", lowerOrganizationName))
	peer, err := peer.New(
		organization,
		peerDirectory,
		int32(apiPort),
		fmt.Sprintf("grpc://%speer-api.%s:%d", lowerOrganizationName, f.config.Domain, f.config.Port),
		int32(chaincodePort),
		fmt.Sprintf("grpc://%speer-chaincode.%s:%d", lowerOrganizationName, f.config.Domain, f.config.Port),
		int32(operationsPort),
		fmt.Sprintf("http://%speer-operations.%s:%d", lowerOrganizationName, f.config.Domain, f.config.Port),
	)
	if err != nil {
		return err
	}
	err = peer.Start()
	if err != nil {
		return err
	}
	f.Lock()
	defer f.Unlock()
	f.peers = append(f.peers, peer)
	log.Printf("Created and started peer for endorsing organization %s", organization.Name())
	return nil
}

func (f *Fablet) createChannel(config Channel) (*common.Block, error) {
	log.Printf("Creating channel %s ...", config.Name)
	opts := []channel.Option{}
	for _, endorsingOrganization := range f.endorsingOrganizations {
		found := false
		for _, organizationName := range config.Organizations {
			if endorsingOrganization.Name() == organizationName {
				found = true
				break
			}
		}
		if found {
			opts = append(opts, channel.AddMSPID(endorsingOrganization.MSP().ID()))
		}
	}
	err := channel.CreateChannel(f.orderer, config.Name, opts...)
	if err != nil {
		return nil, err
	}
	var genesisBlock *common.Block
	for {
		genesisBlock, err = blocks.GetGenesisBlock(f.orderer, config.Name)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}
	opts = []channel.Option{}
	for _, peer := range f.peers {
		found := false
		for _, organizationName := range config.Organizations {
			if peer.Organization().Name() == organizationName {
				found = true
				break
			}
		}
		if found {
			opts = append(opts, channel.AddAnchorPeer(peer.MSPID(), peer.Hostname(), peer.Port()))
		}
	}
	err = channel.UpdateChannel(f.orderer, config.Name, opts...)
	if err != nil {
		return nil, err
	}
	log.Printf("Created channel %s", config.Name)
	return genesisBlock, nil
}

func (f *Fablet) createAndJoinChannel(config Channel) error {
	log.Printf("Creating and joining channel %s ...", config.Name)
	genesisBlock, err := f.createChannel(config)
	if err != nil {
		return err
	}
	ctx := context.Background()
	eg, _ := errgroup.WithContext(ctx)
	for i := range f.peers {
		peer := f.peers[i]
		found := false
		for _, organizationName := range config.Organizations {
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

func (f *Fablet) stop() error {
	if f.proxy != nil {
		err := f.proxy.Stop()
		if err != nil {
			return err
		}
		f.proxy = nil
	}
	if f.console != nil {
		err := f.console.Stop()
		if err != nil {
			return err
		}
		f.console = nil
	}
	for _, peer := range f.peers {
		err := peer.Stop()
		if err != nil {
			return err
		}
	}
	f.peers = []*peer.Peer{}
	if f.orderer != nil {
		err := f.orderer.Stop()
		if err != nil {
			return err
		}
		f.orderer = nil
	}
	return nil
}
