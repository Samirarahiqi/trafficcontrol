package client

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
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	"github.com/apache/trafficcontrol/lib/go-tc"
)

// Users gets an array of Users.
// Deprecated: use GetUsers
func (to *Session) Users() ([]tc.User, error) {
	us, _, err := to.GetUsers()
	return us, err
}

// GetUsers returns all users accessible from current user
func (to *Session) GetUsers() ([]tc.User, ReqInf, error) {
	url := apiBase + "/users.json"
	resp, remoteAddr, err := to.request("GET", url, nil)
	reqInf := ReqInf{CacheHitStatus: CacheHitStatusMiss, RemoteAddr: remoteAddr}
	if err != nil {
		return nil, reqInf, err
	}
	defer resp.Body.Close()

	var data tc.UsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, reqInf, err
	}

	return data.Response, reqInf, nil
}

// GetUserCurrent gets information about the current user
func (to *Session) GetUserCurrent() (*tc.UserCurrent, ReqInf, error) {
	url := apiBase + `/user/current`
	resp := tc.UserCurrentResponse{}
	reqInf, err := get(to, url, &resp)
	if err != nil {
		return nil, reqInf, err
	}
	return &resp.Response, reqInf, nil
}

// CreateUser creates a user
func (to *Session) CreateUser(user *tc.User) (tc.Alerts, ReqInf, error) {
	var remoteAddr net.Addr
	reqBody, err := json.Marshal(user)
	reqInf := ReqInf{CacheHitStatus: CacheHitStatusMiss, RemoteAddr: remoteAddr}
	if err != nil {
		return tc.Alerts{}, reqInf, err
	}
	url := apiBase + "/users.json"
	resp, remoteAddr, err := to.request(http.MethodPost, url, reqBody)
	if err != nil {
		return tc.Alerts{}, reqInf, err
	}
	defer resp.Body.Close()
	var alerts tc.Alerts
	err = json.NewDecoder(resp.Body).Decode(&alerts)
	return alerts, reqInf, nil
}

// UpdateUserByID updates user with the given id
func (to *Session) UpdateUserByID(id int, u tc.User) (tc.Alerts, ReqInf, error) {

	var remoteAddr net.Addr
	reqBody, err := json.Marshal(u)
	reqInf := ReqInf{CacheHitStatus: CacheHitStatusMiss, RemoteAddr: remoteAddr}
	if err != nil {
		return tc.Alerts{}, reqInf, err
	}
	route := apiBase + "/users.json" + strconv.Itoa(id)
	resp, remoteAddr, err := to.request(http.MethodPut, route, reqBody)
	if err != nil {
		return tc.Alerts{}, reqInf, err
	}
	defer resp.Body.Close()
	var alerts tc.Alerts
	err = json.NewDecoder(resp.Body).Decode(&alerts)
	return alerts, reqInf, nil
}

// DeleteUserByID updates user with the given id
func (to *Session) DeleteUserByID(id int) (tc.Alerts, ReqInf, error) {
	route := apiBase + "/users.json" + strconv.Itoa(id)
	resp, remoteAddr, err := to.request(http.MethodDelete, route, nil)
	reqInf := ReqInf{CacheHitStatus: CacheHitStatusMiss, RemoteAddr: remoteAddr}
	if err != nil {
		return tc.Alerts{}, reqInf, err
	}
	defer resp.Body.Close()
	var alerts tc.Alerts
	err = json.NewDecoder(resp.Body).Decode(&alerts)
	return alerts, reqInf, nil
}
