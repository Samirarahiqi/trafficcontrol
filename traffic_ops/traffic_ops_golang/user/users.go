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
	"fmt"
	"strconv"

	"github.com/apache/incubator-trafficcontrol/lib/go-log"
	"github.com/apache/incubator-trafficcontrol/lib/go-tc"
	"github.com/apache/incubator-trafficcontrol/traffic_ops/traffic_ops_golang/api"
	"github.com/apache/incubator-trafficcontrol/traffic_ops/traffic_ops_golang/auth"
	"github.com/apache/incubator-trafficcontrol/traffic_ops/traffic_ops_golang/dbhelpers"
	"github.com/apache/incubator-trafficcontrol/traffic_ops/traffic_ops_golang/tenant"
	"github.com/apache/incubator-trafficcontrol/traffic_ops/traffic_ops_golang/tovalidate"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

//TOUser defines a type alias to define functions on
type TOUser tc.UserNullable

//the refType is passed into the handlers where a copy of its type is used to decode the json.
var refType = TOUser{}

// GetRefType is part of the Identifier interface
func GetRefType() *TOUser {
	return &refType
}

// GetID is part of the Identifier interface
func (user TOUser) GetID() (int, bool) {
	if user.ID == nil {
		return 0, false
	}
	return *user.ID, true
}

// GetAuditName is part of the Identifier interface
func (user TOUser) GetAuditName() string {
	if user.Username != nil {
		return *user.Username
	}
	id, _ := user.GetID()
	return strconv.Itoa(id)
}

// GetType is part of the Identifier interface
func (user TOUser) GetType() string {
	return "user"
}

// SetID is part of the Identifier interface
func (user *TOUser) SetID(i int) {
	user.ID = &i
}

// IsTenantAuthorized implements the Tenantable interface
func (user *TOUser) IsTenantAuthorized(currentUser auth.CurrentUser, db *sqlx.DB) (bool, error) {
	if user == nil || user.TenantID == nil {
		return false, nil
	}
	log.Debugf("IsTenantAuthorized: %d %++v", *user.TenantID, currentUser)
	return tenant.IsResourceAuthorizedToUser(*user.TenantID, currentUser, db)
}

// Validate fulfills the api.Validator interface
func (user TOUser) Validate(db *sqlx.DB) []error {
	matchesID := func(id int) bool { return user.ID == nil || id == *user.ID }

	// consider username and email as invalid passwords
	isGoodPassword := validation.NewStringRule(tovalidate.IsGoodPassword(user.Username, user.Email), "password is too common")
	isUniqueUsername := tovalidate.NewDBUniqueRule(db, "tm_user", "username", matchesID)
	isUniqueEmail := tovalidate.NewDBUniqueRule(db, "tm_user", "email", matchesID)
	roleExists := tovalidate.NewDBExistsRule(db, "role", "id")
	tenantExists := tovalidate.NewDBExistsRule(db, "tenant", "id")

	errs := validation.Errors{
		// TODO: should any character be legal in a username?  Previous versions did not enforce any
		"email":       validation.Validate(user.Email, validation.Required, is.Email, isUniqueEmail),
		"fullName":    validation.Validate(user.FullName, validation.Required),
		"localPasswd": validation.Validate(user.LocalPasswd, validation.Required, isGoodPassword),
		"role":        validation.Validate(user.Role, validation.Required, roleExists),
		"username":    validation.Validate(user.Username, validation.Required, isUniqueUsername),
		"tenant":      validation.Validate(user.TenantID, validation.Required, tenantExists),
	}
	return tovalidate.ToErrors(errs)
}

// Create fulfills the Creator interface
//all implementations of Creator should use transactions and return the proper errorType
//ParsePQUniqueConstraintError is used to determine if a user with conflicting values exists
//if so, it will return an errorType of DataConflict and the type should be appended to the
//generic error message returned
//The insert sql returns the id and lastUpdated values of the newly inserted user and have
//to be added to the struct
func (user *TOUser) Create(db *sqlx.DB, u auth.CurrentUser) (error, tc.ApiErrorType) {
	rollbackTransaction := true
	tx, err := db.Beginx()
	defer func() {
		if tx == nil || !rollbackTransaction {
			return
		}
		err := tx.Rollback()
		if err != nil {
			log.Errorln(errors.New("rolling back transaction: " + err.Error()))
		}
	}()

	if err != nil {
		log.Error.Printf("could not begin transaction: %v", err)
		return err, tc.SystemError
	}

	// encrypt password before storing the user
	encPassword, err := auth.DerivePassword(*user.LocalPasswd)
	if err != nil {
		log.Errorf("could not encrypt password: %v", err)
		return err, tc.SystemError
	}
	user.LocalPasswd = &encPassword

	resultRows, err := tx.NamedQuery(insertQuery(), user)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			err, eType := dbhelpers.ParsePQUniqueConstraintError(pqErr)
			if eType == tc.DataConflictError {
				return errors.New("a user with " + err.Error()), eType
			}
			return err, eType
		}
		log.Errorf("received non pq error: %++v from create execution", err)
		return err, tc.SystemError
	}
	defer resultRows.Close()

	var id int
	var lastUpdated tc.TimeNoMod
	rowsAffected := 0
	for resultRows.Next() {
		rowsAffected++
		if err := resultRows.Scan(&id, &lastUpdated); err != nil {
			log.Error.Printf("could not scan id from insert: %s\n", err)
			return err, tc.SystemError
		}
	}
	if rowsAffected == 0 {
		err = errors.New("no user was inserted, no id was returned")
		log.Errorln(err)
		return err, tc.SystemError
	}
	if rowsAffected > 1 {
		err = errors.New("too many ids returned from user insert")
		log.Errorln(err)
		return err, tc.SystemError
	}
	user.SetID(id)
	user.LastUpdated = &lastUpdated
	err = tx.Commit()
	if err != nil {
		log.Errorln("Could not commit transaction: ", err)
		return err, tc.SystemError
	}
	rollbackTransaction = false
	return nil, tc.NoError
}

// Read fulfills the Reader interface
func (user *TOUser) Read(db *sqlx.DB, parameters map[string]string, u auth.CurrentUser) ([]interface{}, []error, tc.ApiErrorType) {
	var rows *sqlx.Rows

	// Query Parameters to Database Query column mappings
	// see the fields mapped in the SQL query
	queryParamsToQueryCols := map[string]dbhelpers.WhereColumnInfo{
		"active":       dbhelpers.WhereColumnInfo{Column: "u.active", Checker: nil},
		"company":      dbhelpers.WhereColumnInfo{Column: "u.company", Checker: nil},
		"email":        dbhelpers.WhereColumnInfo{Column: "u.email", Checker: nil},
		"fullName":     dbhelpers.WhereColumnInfo{Column: "u.full_name", Checker: nil},
		"gid":          dbhelpers.WhereColumnInfo{Column: "u.gid", Checker: api.IsInt},
		"id":           dbhelpers.WhereColumnInfo{Column: "u.id", Checker: api.IsInt},
		"lastUpdated":  dbhelpers.WhereColumnInfo{Column: "u.last_updated", Checker: nil},
		"newUser":      dbhelpers.WhereColumnInfo{Column: "u.new_user", Checker: nil},
		"publicSSHKey": dbhelpers.WhereColumnInfo{Column: "u.public_ssh_key", Checker: nil},
		"role":         dbhelpers.WhereColumnInfo{Column: "u.role", Checker: api.IsInt},
		"rolename":     dbhelpers.WhereColumnInfo{Column: "r.name", Checker: nil},
		"uid":          dbhelpers.WhereColumnInfo{Column: "u.uid", Checker: api.IsInt},
		"username":     dbhelpers.WhereColumnInfo{Column: "u.username", Checker: nil},
	}
	where, orderBy, queryValues, errs := dbhelpers.BuildWhereAndOrderBy(parameters, queryParamsToQueryCols)
	if len(errs) > 0 {
		return nil, errs, tc.DataConflictError
	}

	query := selectQuery() + where + orderBy
	log.Debugln("Query is ", query)

	rows, err := db.NamedQuery(query, queryValues)
	if err != nil {
		log.Errorf("Error querying Users: %v", err)
		return nil, []error{tc.DBError}, tc.SystemError
	}
	defer rows.Close()

	users := []interface{}{}
	for rows.Next() {
		var s TOUser
		if err = rows.StructScan(&s); err != nil {
			log.Errorf("error parsing User rows: %v", err)
			return nil, []error{err}, tc.SystemError
		}
		users = append(users, s)
	}

	return users, []error{}, tc.NoError
}

// Update fulfills the Updater interface
//all implementations of Updater should use transactions and return the proper errorType
//ParsePQUniqueConstraintError is used to determine if a user with conflicting values exists
//if so, it will return an errorType of DataConflict and the type should be appended to the
//generic error message returned
func (user *TOUser) Update(db *sqlx.DB, u auth.CurrentUser) (error, tc.ApiErrorType) {
	rollbackTransaction := true
	tx, err := db.Beginx()
	defer func() {
		if tx == nil || !rollbackTransaction {
			return
		}
		err := tx.Rollback()
		if err != nil {
			log.Errorln(errors.New("rolling back transaction: " + err.Error()))
		}
	}()

	if err != nil {
		log.Error.Printf("could not begin transaction: %v", err)
		return err, tc.SystemError
	}

	// encrypt password before storing the user
	encPassword, err := auth.DerivePassword(*user.LocalPasswd)
	if err != nil {
		log.Errorf("could not encrypt password: %v", err)
		return err, tc.SystemError
	}
	user.LocalPasswd = &encPassword

	log.Debugf("about to run exec query: %s with user: %++v", updateQuery(), user)
	resultRows, err := tx.NamedQuery(updateQuery(), user)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			err, eType := dbhelpers.ParsePQUniqueConstraintError(pqErr)
			if eType == tc.DataConflictError {
				return errors.New("a user with " + err.Error()), eType
			}
			return err, eType
		}
		log.Errorf("received error: %++v from update execution", err)
		return err, tc.SystemError
	}
	defer resultRows.Close()

	var lastUpdated tc.TimeNoMod
	rowsAffected := 0
	for resultRows.Next() {
		rowsAffected++
		if err := resultRows.Scan(&lastUpdated); err != nil {
			log.Error.Printf("could not scan lastUpdated from insert: %s\n", err)
			return err, tc.SystemError
		}
	}
	log.Debugf("lastUpdated: %++v", lastUpdated)
	user.LastUpdated = &lastUpdated
	if rowsAffected != 1 {
		if rowsAffected < 1 {
			return errors.New("no user found with this id"), tc.DataMissingError
		}
		return fmt.Errorf("this update affected too many rows: %d", rowsAffected), tc.SystemError
	}
	err = tx.Commit()
	if err != nil {
		log.Errorln("Could not commit transaction: ", err)
		return err, tc.SystemError
	}
	rollbackTransaction = false
	return nil, tc.NoError
}

// Delete fulfills the Deleter interface
//all implementations of Deleter should use transactions and return the proper errorType
// For user, delete does not actually delete from the table,  it only sets the "active" flag to false.
// That way the integrity of the db is maintained where there is a history for this user.
func (user *TOUser) Delete(db *sqlx.DB, u auth.CurrentUser) (error, tc.ApiErrorType) {
	rollbackTransaction := true
	tx, err := db.Beginx()
	defer func() {
		if tx == nil || !rollbackTransaction {
			return
		}
		err := tx.Rollback()
		if err != nil {
			log.Errorln(errors.New("rolling back transaction: " + err.Error()))
		}
	}()

	if err != nil {
		log.Error.Printf("could not begin transaction: %v", err)
		return err, tc.SystemError
	}
	log.Debugf("about to run exec query: %s with user: %++v", deactivateUserQuery(), user)
	result, err := tx.NamedExec(deactivateUserQuery(), user)
	if err != nil {
		log.Errorf("received error: %++v from delete execution", err)
		return err, tc.SystemError
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err, tc.SystemError
	}
	if rowsAffected != 1 {
		if rowsAffected < 1 {
			return errors.New("no user with that id found"), tc.DataMissingError
		}
		return fmt.Errorf("this create affected too many rows: %d", rowsAffected), tc.SystemError
	}
	err = tx.Commit()
	if err != nil {
		log.Errorln("Could not commit transaction: ", err)
		return err, tc.SystemError
	}
	rollbackTransaction = false
	return nil, tc.NoError
}

func selectQuery() string {
	query := `SELECT
u.active,
u.address_line1,
u.address_line2,
u.city,
u.company,
u.country,
u.email,
u.full_name,
u.gid,
u.id,
u.last_updated,
u.local_passwd,
u.new_user,
u.phone_number,
u.postal_code,
u.public_ssh_key,
u.registration_sent,
u.role,
r.name AS role_name,
u.state_or_province,
u.tenant_id,
u.token,
u.uid,
u.username

FROM tm_user u JOIN role r ON u.role = r.id`
	return query
}

func updateQuery() string {
	query := `UPDATE
tm_user SET
active=:active,
address_line1=:address_line1,
address_line2=:address_line2,
city=:city,
company=:company,
country=:country,
email=:email,
full_name=:full_name,
gid=:gid,
id=:id,
last_updated=:last_updated,
local_passwd=:local_passwd,
new_user=:new_user,
phone_number=:phone_number,
postal_code=:postal_code,
public_ssh_key=:public_ssh_key,
registration_sent=:registration_sent,
role=:role,
state_or_province=:state_or_province,
tenant_id=:tenant_id,
token=:token,
uid=:uid,
username=:username

WHERE id=:id RETURNING last_updated`
	return query
}

func insertQuery() string {
	query := `INSERT INTO tm_user (
active,
address_line1,
address_line2,
city,
company,
country,
email,
full_name,
gid,
id,
last_updated,
local_passwd,
new_user,
phone_number,
postal_code,
public_ssh_key,
registration_sent,
role,
state_or_province,
tenant_id,
token,
uid,
username
) VALUES (
:active,
:address_line1,
:address_line2,
:city,
:company,
:country,
:email,
:full_name,
:gid,
:id,
:last_updated,
:local_passwd,
:new_user,
:phone_number,
:postal_code,
:public_ssh_key,
:registration_sent,
:role,
:state_or_province,
:tenant_id,
:token,
:uid,
:username
) RETURNING id,last_updated`
	return query
}

func deactivateUserQuery() string {
	query := `UPDATE tm_user
SET active=false
WHERE id=:id`
	return query
}
