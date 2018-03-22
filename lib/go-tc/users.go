package tc

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

// UsersResponse ...
type UsersResponse struct {
	Response []User `json:"response"`
}

// User contains information about a given user in Traffic Ops.
type User struct {
	Active           bool      `json:"active,omitempty" db:"active"`
	AddressLine1     string    `json:"addressLine1" db:"address_line1"`
	AddressLine2     string    `json:"addressLine2" db:"address_line2"`
	City             string    `json:"city" db:"city"`
	Company          string    `json:"company,omitempty" db:"company"`
	Country          string    `json:"country" db:"country"`
	Email            string    `json:"email,omitempty" db:"email"`
	FullName         string    `json:"fullName,omitempty" db:"full_name"`
	GID              int       `json:"gid,omitempty" db:"gid"`
	ID               int       `json:"id,omitempty" db:"id"`
	LastUpdated      TimeNoMod `json:"lastUpdated,omitempty" db:"last_updated"`
	LocalPasswd      string    `json:"localPasswd" db:"local_passwd"`
	NewUser          bool      `json:"newUser,omitempty" db:"new_user"`
	PhoneNumber      string    `json:"phoneNumber" db:"phone_number"`
	PostalCode       string    `json:"postalCode" db:"postal_code"`
	PublicSSHKey     string    `json:"publicSshKey,omitempty" db:"publicSshKey"`
	RegistrationSent TimeNoMod `json:"registrationSent" db:"registration_sent"`
	Role             int       `json:"role,omitempty" db:"role"`
	RoleName         string    `json:"rolename,omitempty"`
	StateOrProvince  string    `json:"stateOrProvince" db:"state_or_province"`
	TenantID         int       `json:"tentantId" db:"tenant_id"`
	Token            string    `json:"token" db:"token"`
	UID              int       `json:"uid,omitempty" db:"uid"`
	Username         string    `json:"username,omitempty" db:"username"`
}

// UserNullable contains information about a given user in Traffic Ops.
type UserNullable struct {
	Active           *bool      `json:"active,omitempty" db:"active"`
	AddressLine1     *string    `json:"addressLine1" db:"address_line1"`
	AddressLine2     *string    `json:"addressLine2" db:"address_line2"`
	City             *string    `json:"city" db:"city"`
	Company          *string    `json:"company,omitempty" db:"company"`
	Country          *string    `json:"country" db:"country"`
	Email            *string    `json:"email,omitempty" db:"email"`
	FullName         *string    `json:"fullName,omitempty" db:"full_name"`
	GID              *int       `json:"gid,omitempty" db:"gid"`
	ID               *int       `json:"id,omitempty" db:"id"`
	LastUpdated      *TimeNoMod `json:"lastUpdated,omitempty" db:"last_updated"`
	LocalPasswd      *string    `json:"localPasswd" db:"local_passwd"`
	NewUser          *bool      `json:"newUser,omitempty" db:"new_user"`
	PhoneNumber      *string    `json:"phoneNumber" db:"phone_number"`
	PostalCode       *string    `json:"postalCode" db:"postal_code"`
	PublicSSHKey     *string    `json:"publicSshKey,omitempty" db:"public_ssh_key"`
	RegistrationSent *TimeNoMod `json:"registrationSent" db:"registration_sent"`
	Role             *int       `json:"role,omitempty" db:"role"`
	RoleName         *string    `json:"rolename,omitempty" db:"role_name"`
	StateOrProvince  *string    `json:"stateOrProvince" db:"state_or_province"`
	TenantID         *int       `json:"tentantId" db:"tenant_id"`
	Token            *string    `json:"token" db:"token"`
	UID              *int       `json:"uid,omitempty" db:"uid"`
	Username         *string    `json:"username,omitempty" db:"username"`
}

// UserCredentials contains Traffic Ops login credentials
type UserCredentials struct {
	Username string `json:"u"`
	Password string `json:"p"`
}
