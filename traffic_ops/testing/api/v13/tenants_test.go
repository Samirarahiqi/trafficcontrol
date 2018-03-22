package v13

/*

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"testing"

	"github.com/apache/incubator-trafficcontrol/lib/go-log"
	tc "github.com/apache/incubator-trafficcontrol/lib/go-tc"
)

func TestTenants(t *testing.T) {

	CreateTestTenants(t)
	//UpdateTestTenants(t)
	//GetTestTenants(t)
	//DeleteTestTenants(t)
	//GetTestActiveTenants(t)
}

func CreateTestTenants(t *testing.T) {

	for _, ten := range testData.Tenants {
		resp, _, err := TOSession.CreateTenant(ten)
		log.Debugln("Response: ", resp)
		if err != nil {
			t.Errorf("could not CREATE users: %v\n", err)
		}
	}

}

func UpdateTestTenants(t *testing.T) {

	firstTenant := testData.Tenants[0]
	// Retrieve the Tenant by name so we can get the id for the Update
	resp, _, err := TOSession.GetTenantByName(firstTenant.Name)
	if err != nil {
		t.Errorf("cannot GET Tenant by name: %v - %v\n", firstTenant.Name, err)
	}
	remoteTenant := resp[0]
	expectedName := "testTenant1"
	remoteTenant.Name = expectedName
	var alert tc.Alerts
	alert, _, err = TOSession.UpdateTenantByID(remoteTenant.ID, remoteTenant)
	if err != nil {
		t.Errorf("cannot UPDATE Tenant by id: %v - %v\n", err, alert)
	}

	// Retrieve the Tenant to check Tenant name got updated
	resp, _, err = TOSession.GetTenantByID(remoteTenant.ID)
	if err != nil {
		t.Errorf("cannot GET Tenant by name: %v - %v\n", firstTenant.Name, err)
	}
	respTenant := resp[0]
	if respTenant.Name != expectedName {
		t.Errorf("results do not match actual: %s, expected: %s\n", respTenant.Name, expectedName)
	}

}

func GetTestTenants(t *testing.T) {

	for _, ten := range testData.Tenants {
		resp, _, err := TOSession.GetTenantByName(ten.Name)
		if err != nil {
			t.Errorf("cannot GET Tenant by name: %v - %v\n", err, resp)
		}
	}
}

func DeleteTestTenants(t *testing.T) {

	for _, ten := range testData.Tenants {
		// Retrieve the Tenant by name so we can get the id for the Update
		resp, _, err := TOSession.GetTenantByName(ten.Name)
		if err != nil {
			t.Errorf("cannot GET Tenant by name: %v - %v\n", ten.Name, err)
		}
		respTenant := resp[0]

		delResp, _, err := TOSession.DeleteTenantByID(respTenant.ID)
		if err != nil {
			t.Errorf("cannot DELETE Tenant by name: %v - %v\n", err, delResp)
		}

		// Retrieve the Tenant to see if it got deleted
		Tenants, _, err := TOSession.GetTenantByName(ten.Name)
		if err != nil {
			t.Errorf("error deleting Tenant name: %s\n", err.Error())
		}
		if len(Tenants) > 0 {
			t.Errorf("expected Tenant name: %s to be deleted\n", ten.Name)
		}
	}
}
