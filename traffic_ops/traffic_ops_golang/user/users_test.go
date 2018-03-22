package user

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/apache/incubator-trafficcontrol/lib/go-tc"
	"github.com/apache/incubator-trafficcontrol/traffic_ops/traffic_ops_golang/api"
	"github.com/apache/incubator-trafficcontrol/traffic_ops/traffic_ops_golang/auth"
	"github.com/apache/incubator-trafficcontrol/traffic_ops/traffic_ops_golang/test"
	"github.com/jmoiron/sqlx"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func getTestUsers() []TOUser {
	active := true
	company := "Yoyodyne"
	email := "buckaroo@yoyodyne.bz"
	fullName := "Buckaroo Bonzai"
	id := 0
	lastUpdated := tc.TimeNoMod{Time: time.Now(), Valid: true}
	localPasswd := "letmein"
	newUser := false
	registrationSent := tc.TimeNoMod{}
	role := 0
	roleName := "Skipper"
	tenantID := 1
	uID := 0
	username := "buckaroo"
	empty := ""

	testUser := TOUser{
		Active:           &active,
		AddressLine1:     &empty,
		AddressLine2:     &empty,
		City:             &empty,
		Company:          &company,
		Country:          &empty,
		Email:            &email,
		FullName:         &fullName,
		GID:              &id,
		ID:               &id,
		LastUpdated:      &lastUpdated,
		LocalPasswd:      &localPasswd,
		NewUser:          &newUser,
		PhoneNumber:      &empty,
		PostalCode:       &empty,
		PublicSSHKey:     &empty,
		RegistrationSent: &registrationSent,
		Role:             &role,
		RoleName:         &roleName,
		StateOrProvince:  &empty,
		TenantID:         &tenantID,
		Token:            &empty,
		UID:              &uID,
		Username:         &username,
	}

	testUser2 := testUser
	username2 := "penny"
	role2 := 1
	testUser2.Username = &username2
	testUser2.Email = &username2
	testUser2.Role = &role2
	return []TOUser{testUser, testUser2}
}

func TestReadUsers(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	defer db.Close()

	refType := GetRefType()

	testUsers := getTestUsers()
	cols := test.ColsFromStructByTag("db", TOUser{})
	t.Logf("Columns: %v", cols)
	rows := sqlmock.NewRows(cols)
	for _, ts := range testUsers {
		rows = rows.AddRow(
			*ts.Active,
			*ts.AddressLine1,
			*ts.AddressLine2,
			*ts.City,
			*ts.Company,
			*ts.Country,
			*ts.Email,
			*ts.FullName,
			*ts.GID,
			*ts.ID,
			*ts.LastUpdated,
			*ts.LocalPasswd,
			*ts.NewUser,
			*ts.PhoneNumber,
			*ts.PostalCode,
			*ts.PublicSSHKey,
			*ts.RegistrationSent,
			*ts.Role,
			*ts.RoleName,
			*ts.StateOrProvince,
			*ts.TenantID,
			*ts.Token,
			*ts.UID,
			*ts.Username,
		)
	}
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	v := map[string]string{}

	users, errs, _ := refType.Read(db, v, auth.CurrentUser{})
	if len(errs) > 0 {
		t.Errorf("user.Read expected: no errors, actual: %v", errs)
	}

	if len(users) != 2 {
		t.Errorf("user.Read expected: len(users) == 2, actual: %v", len(users))
	}
}

func TestIdentifier(t *testing.T) {
	u := TOUser{}
	id, valid := u.GetID()
	// no id,  so GetID should return valid==false
	if id != 0 || valid {
		t.Errorf("expected GetID to return 0, false, got %d %v", id, valid)
	}
	if u.GetAuditName() != "0" {
		t.Errorf("expected GetAuditName to return %s,  got %s", "0", u.GetAuditName())
	}

	// set id
	u.ID = &id
	id, valid = u.GetID()
	if id != 0 || !valid {
		t.Errorf("expected GetID to return 0, true, got %d %v", id, valid)
	}
	username := "0"
	if u.GetAuditName() != username {
		t.Errorf("expected GetAuditName to return %s,  got %s", username, u.GetAuditName())
	}

	// set username
	username = "buckaroo"
	u.Username = &username
	if u.GetAuditName() != username {
		t.Errorf("expected GetAuditName to return %s,  got %s", username, u.GetAuditName())
	}
	if u.GetType() != "user" {
		t.Errorf("expected GetType to return %s,  got %s", "user", u.GetType())
	}
}

func TestFuncs(t *testing.T) {
	if strings.Index(selectQuery(), "SELECT") != 0 {
		t.Errorf("expected selectQuery to start with SELECT")
	}
	if strings.Index(insertQuery(), "INSERT") != 0 {
		t.Errorf("expected insertQuery to start with INSERT")
	}
	if strings.Index(updateQuery(), "UPDATE") != 0 {
		t.Errorf("expected updateQuery to start with UPDATE")
	}

	// no DELETE -- just UPDATE setting active to false
	if strings.Index(deactivateUserQuery(), "UPDATE") != 0 {
		t.Errorf("expected deactivateUserQuery to start with UPDATE")
	}
}

func TestInterfaces(t *testing.T) {
	var i interface{}
	i = &TOUser{}

	if _, ok := i.(api.Creator); !ok {
		t.Errorf("user must be creator")
	}
	if _, ok := i.(api.Reader); !ok {
		t.Errorf("user must be reader")
	}
	if _, ok := i.(api.Updater); !ok {
		t.Errorf("user must be updater")
	}
	if _, ok := i.(api.Deleter); !ok {
		t.Errorf("user must be deleter")
	}
	if _, ok := i.(api.Identifier); !ok {
		t.Errorf("user must be Identifier")
	}
	if _, ok := i.(api.Tenantable); !ok {
		t.Errorf("user must be Tenantable")
	}
}

func TestValidate(t *testing.T) {
	n := "validuser"
	c := TOUser{Username: &n}
	errs := test.SortErrors(c.Validate(nil))
	expectedErrs := test.SortErrors([]error{
		errors.New(`'email' cannot be blank`),
		errors.New(`'localPasswd' cannot be blank`),
		errors.New(`'fullName' cannot be blank`),
		errors.New(`'role' cannot be blank`),
		errors.New(`'tenant' cannot be blank`),
	})

	if !reflect.DeepEqual(expectedErrs, errs) {
		t.Errorf("expected %++v, got %++v", expectedErrs, errs)
	}

	testUsers := getTestUsers()
	errs = test.SortErrors(testUsers[0].Validate(nil))

	expectedErrs = test.SortErrors([]error{
		errors.New(`'localPasswd' password is too common`),
		errors.New(`'role' cannot be blank`),
	})

	if !reflect.DeepEqual(expectedErrs, errs) {
		t.Errorf("expected %++v, got %++v", expectedErrs, errs)
	}

	errs = test.SortErrors(testUsers[1].Validate(nil))

	expectedErrs = []error{
		errors.New(`'email' must be a valid email address`),
		errors.New(`'localPasswd' password is too common`),
	}

	if !reflect.DeepEqual(expectedErrs, errs) {
		t.Errorf("expected %++v, got %++v", expectedErrs, errs)
	}

}
