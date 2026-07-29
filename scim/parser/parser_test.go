package parser_test // ScimFilter

import (
	"regexp"
	"strings"
	"testing"

	"github.com/arturoeanton/goscim/scim/parser"
)

/*
filter=userName eq "bjensen"
filter=name.familyName co "O'Malley"
filter=userName sw "J"
filter=title pr
filter=meta.lastModified gt "2011-05-13T04:42:34Z"
filter=meta.lastModified ge "2011-05-13T04:42:34Z"
filter=meta.lastModified lt "2011-05-13T04:42:34Z"
filter=meta.lastModified le "2011-05-13T04:42:34Z"
filter=title pr and userType eq "Employee"
filter=title pr or userType eq "Intern"
filter=userType eq "Employee" and (emails co "example.com" or emails co "example.org")
*/
func TestParser(t *testing.T) {
	query, err := parser.FilterToN1QL("User", "urn:ietf:params:scim:schemas:extension:one:2.0:User.userType eq \"Employee\" and (emails sw \"example.com\" or a.a.emails sw \"example.org\")", nil)
	if err != nil {
		t.Fatal(err)
	}
	want := "SELECT * FROM `User` WHERE `urn:ietf:params:scim:schemas:extension:one:2.0:User`.`userType` = $1 and (`emails` LIKE $2 ESCAPE \"\\\\\" or `a`.`a`.`emails` LIKE $3 ESCAPE \"\\\\\")"
	if query.Page != want {
		t.Errorf("Query is %s", query.Page)
	}
}

func TestParser2(t *testing.T) {
	query, err := parser.FilterToN1QL("User", "urn:ietf:params:scim:schemas:extension:one:2.0:Element.boolean eq true", nil)
	if err != nil {
		t.Fatal(err)
	}
	if query.Page != "SELECT * FROM `User` WHERE `urn:ietf:params:scim:schemas:extension:one:2.0:Element`.`boolean` = $1" {
		t.Errorf("Query is %s", query.Page)
	}
	if len(query.Params) != 1 || query.Params[0] != true {
		t.Errorf("Params = %#v", query.Params)
	}
}

func TestRex(t *testing.T) {
	sample := "urn:ietf:params:scim:schemas:extension:3.3:one:2.0:Element.boolean"
	re := regexp.MustCompile(`^(urn[:\w\.\_]*)(:-*)?(:[\w]*)(\.)(.*)$`)
	urn := ""
	if re.MatchString(sample) {
		urn = "`" + re.ReplaceAllString(sample, `${1}${2}${3}`) + "`."
	}
	path := re.ReplaceAllString(sample, `${5}`)

	path = urn + "`" + strings.Join(strings.Split(path, "."), "`.`") + "`"

	if path != "`urn:ietf:params:scim:schemas:extension:3.3:one:2.0:Element`.`boolean`" {
		t.Error("-------")
		t.Error(urn)
		t.Error(path)
		t.Error("-------")
	}
}
