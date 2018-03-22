// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tovalidate

import (
	"bufio"
	"os"
	"strings"

	"github.com/apache/incubator-trafficcontrol/lib/go-log"
)

// NoSpaces returns true if the string has no spaces
func NoSpaces(str string) bool {
	return !strings.ContainsAny(str, " ")
}

// NoPeriods returns true if the string has no periods
func NoPeriods(str string) bool {
	return !strings.ContainsAny(str, ".")
}

// IsOneOfString generates a validator function returning whether string is in the set of strings
func IsOneOfString(set ...string) func(string) bool {
	return func(s string) bool {
		for _, x := range set {
			if s == x {
				return true
			}
		}
		return false
	}
}

// IsOneOfStringICase is a case-insensitive version of IsOneOfString
func IsOneOfStringICase(set ...string) func(string) bool {
	var lowcased []string
	for _, s := range set {
		lowcased = append(lowcased, strings.ToLower(s))
	}
	return IsOneOfString(lowcased...)
}

// IsGoodPassword returns a func that checks given string against a list of bad passwords. Caller should
// pass in a list of words from the context of the caller (e.g. username and email). The list of common passwords
// is read from a fixed configuration file.
func IsGoodPassword(badPasswords ...*string) func(string) bool {
	return func(newPassword string) bool {
		if len(newPassword) < 8 {
			return false
		}
		for _, pw := range badPasswords {
			if pw == nil {
				continue
			}
			if *pw == newPassword {
				return false
			}
		}

		// TODO: we consider that checking new passwords does not happen all that often.
		// So, we just read this file when needed and do not retain the data.
		fn := os.Getenv("TO_DIR")
		if len(fn) == 0 {
			fn = "/opt/traffic_ops/app"
		}
		fn += "/conf/invalid_passwords.txt"
		f, err := os.Open(fn)
		if err != nil {
			log.Errorf("unable to read '%s' to check for invalid passwords", fn)
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if newPassword == scanner.Text() {
				return false
			}
		}
		return true
	}
}
