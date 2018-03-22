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

func TestUsers(t *testing.T) {

	CreateTestUsers(t)
	UpdateTestUsers(t)
	//GetTestUsers(t)
	//DeleteTestUsers(t)
	//GetTestActiveUsers(t)
}

func CreateTestUsers(t *testing.T) {

	for _, u := range testData.Users {
		resp, _, err := TOSession.CreateUser(u)
		log.Debugln("Response: ", resp)
		if err != nil {
			t.Errorf("could not CREATE users: %v\n", err)
		}
	}

}

func UpdateTestUsers(t *testing.T) {

	firstUser := testData.Users[0]
	// Retrieve the User by name so we can get the id for the Update
	resp, _, err := TOSession.GetUserByUsername(firstUser.Username)
	if err != nil {
		t.Errorf("cannot GET User by name: %v - %v\n", firstUser.Username, err)
	}
	remoteUser := resp[0]
	expectedUsername := "testUser1"
	remoteUser.Username = expectedUsername
	var alert tc.Alerts
	alert, _, err = TOSession.UpdateUserByID(remoteUser.ID, remoteUser)
	if err != nil {
		t.Errorf("cannot UPDATE User by id: %v - %v\n", err, alert)
	}

	// Retrieve the User to check User name got updated
	resp, _, err = TOSession.GetUserByID(remoteUser.ID)
	if err != nil {
		t.Errorf("cannot GET User by name: %v - %v\n", firstUser.Username, err)
	}
	respUser := resp[0]
	if respUser.Username != expectedUsername {
		t.Errorf("results do not match actual: %s, expected: %s\n", respUser.Username, expectedUsername)
	}

}

func GetTestUsers(t *testing.T) {

	for _, u := range testData.Users {
		resp, _, err := TOSession.GetUserByUsername(u.Username)
		if err != nil {
			t.Errorf("cannot GET User by name: %v - %v\n", err, resp)
		}
	}
}

func DeleteTestUsers(t *testing.T) {

	for _, u := range testData.Users {
		// Retrieve the User by name so we can get the id for the Update
		resp, _, err := TOSession.GetUserByUsername(u.Username)
		if err != nil {
			t.Errorf("cannot GET User by name: %v - %v\n", u.Username, err)
		}
		respUser := resp[0]

		delResp, _, err := TOSession.DeleteUserByID(respUser.ID)
		if err != nil {
			t.Errorf("cannot DELETE User by name: %v - %v\n", err, delResp)
		}

		// Retrieve the User to see if it got deleted
		Users, _, err := TOSession.GetUserByUsername(u.Username)
		if err != nil {
			t.Errorf("error deleting User name: %s\n", err.Error())
		}
		if len(Users) > 0 {
			t.Errorf("expected User name: %s to be deleted\n", u.Username)
		}
	}
}
