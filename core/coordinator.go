package core

import (
	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/util"
)

type vehicleCoordinator struct {
	tracked map[api.Vehicle]interface{}
}

var coordinator *vehicleCoordinator

func init() {
	coordinator = &vehicleCoordinator{
		tracked: make(map[api.Vehicle]interface{}),
	}
}

func (lp *vehicleCoordinator) acquire(owner interface{}, vehicle api.Vehicle) {
	lp.tracked[vehicle] = owner
}

func (lp *vehicleCoordinator) release(vehicle api.Vehicle) {
	delete(lp.tracked, vehicle)
}

func (lp *vehicleCoordinator) availableVehicles(owner interface{}, vehicles []api.Vehicle) []api.Vehicle {
	var res []api.Vehicle

	for _, vv := range vehicles {
		if _, ok := vv.(api.ChargeState); ok {
			if o, ok := lp.tracked[vv]; o == owner || !ok {
				res = append(res, vv)
			}
		}
	}

	return res
}

// find active vehicle by charge state
func (lp *vehicleCoordinator) identifyVehicleByStatus(log *util.Logger, owner interface{}, vehicles []api.Vehicle) api.Vehicle {
	available := lp.availableVehicles(owner, vehicles)

	var res api.Vehicle
	for _, vehicle := range available {
		if vs, ok := vehicle.(api.ChargeState); ok {
			status, err := vs.Status()

			if err != nil {
				log.ERROR.Println("vehicle status:", err)
				continue
			}

			log.DEBUG.Printf("vehicle status: %s (%s)", status, vehicle.Title())

			// vehicle is plugged or charging, so it should be the right one
			if status == api.StatusB || status == api.StatusC {
				if res != nil {
					log.WARN.Println("vehicle status: >1 matches, giving up")
					return nil
				}

				res = vehicle
			}
		}
	}

	return res
}
