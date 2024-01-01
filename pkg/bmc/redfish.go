package bmc

import (
	"fmt"

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

func getVirtualMediaCD(vms []*redfish.VirtualMedia) (*redfish.VirtualMedia, error) {
	var vMedia *redfish.VirtualMedia

	for _, vm := range vms {
		if vm.ID == "Cd" {
			vMedia = vm
			break
		}
	}

	if vMedia == nil {
		return nil, fmt.Errorf("CD virtual media not found")
	}

	return vMedia, nil
}

type RedFish struct {
	system *redfish.ComputerSystem
	cd     *redfish.VirtualMedia
	config *gofish.ClientConfig
	client *gofish.APIClient
}

func (rf *RedFish) Boot() error {
	if err := rf.initSystem(); err != nil {
		return fmt.Errorf("failed to init system: %w", err)
	}

	if err := rf.system.Reset(redfish.OnResetType); err != nil {
		return fmt.Errorf("failed to boot system: %v", err)
	}

	return nil
}

func (rf *RedFish) SetBootMedia() error {
	if err := rf.initSystem(); err != nil {
		return fmt.Errorf("failed to init system: %w", err)
	}

	boot := redfish.Boot{
		BootSourceOverrideTarget: redfish.CdBootSourceOverrideTarget,
	}
	if err := rf.system.SetBoot(boot); err != nil {
		return fmt.Errorf("failed to set boot media: %v", err)
	}

	return nil
}

func (rf *RedFish) InsertMedia(url string) error {
	if err := rf.initCD(); err != nil {
		return err
	}

	err := rf.cd.InsertMediaConfig(redfish.VirtualMediaConfig{
		Image:    url,
		Inserted: true,
	})
	if err != nil {
		return fmt.Errorf("failed to insert media: %v", err)
	}
	return nil
}

func (rf *RedFish) initClient() error {
	if rf.client != nil {
		return nil
	}

	var err error

	if rf.client, err = gofish.Connect(*rf.config); err != nil {
		return fmt.Errorf("failed to initialize the RedFish client: %v", err)
	}

	return nil
}

func (rf *RedFish) initConfig(url, username, password string) {
	rf.config = &gofish.ClientConfig{
		Endpoint: url,
		// TODO: (GAL-311) Parametrize
		Insecure:  true,
		Username:  username,
		Password:  password,
		BasicAuth: true,
	}
}

func (rf *RedFish) initSystem() error {
	if rf.system != nil {
		return nil
	}
	if rf.client == nil {
		return fmt.Errorf("client not initialized")
	}

	systems, err := rf.client.Service.Systems()
	if err != nil {
		return fmt.Errorf("failed to get the RedFish systems: %v", err)
	}
	if len(systems) == 0 {
		return fmt.Errorf("no systems found")
	}

	rf.system = systems[0]
	return nil
}

func (rf *RedFish) initCD() error {
	if rf.cd != nil {
		return nil
	}

	if err := rf.initSystem(); err != nil {
		return fmt.Errorf("failed to init system: %w", err)
	}

	managerNames := rf.system.ManagedBy
	if len(managerNames) != 1 {
		return fmt.Errorf("only 1 manager is expected for each system")
	}
	managers, err := rf.client.Service.Managers()
	if err != nil {
		return fmt.Errorf("failed to get the RedFish managers: %v", err)
	}

	var manager *redfish.Manager
	for _, m := range managers {
		if m.ODataID == managerNames[0] {
			manager = m
			break
		}
	}

	if manager == nil {
		return fmt.Errorf("manager for the system %s not found", rf.system.Name)
	}

	vMedia, err := manager.VirtualMedia()
	if err != nil {
		return fmt.Errorf("failed to get the RedFish virtual media: %v", err)
	}

	if rf.cd, err = getVirtualMediaCD(vMedia); err != nil {
		return fmt.Errorf("failed to get the RedFish virtual media CD: %v", err)
	}

	return nil
}

func (rf *RedFish) Close() {
	rf.client.Logout()
}

func NewRedFish(url, username, password string) (*RedFish, error) {
	rf := &RedFish{}
	rf.initConfig(url, username, password)
	if err := rf.initClient(); err != nil {
		return nil, err
	}
	return rf, nil
}
