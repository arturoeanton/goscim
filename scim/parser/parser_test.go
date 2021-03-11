package parser_test // ScimFilter

import (
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
	query := parser.FilterToN1QL("User", "urn:ietf:params:scim:schemas:extension:one:2.0:Element.userType eq \"Employee\" and (emails sw \"example.com\" or a.a.emails sw \"example.org\")")
	if query != "SELECT * FROM `User` WHERE `urn:ietf:params:scim:schemas:extension:one:2.0:Element`.userType` = \"Employee\" and (`emails` LIKE \"example.com%\" or `a`.`a`.`emails` LIKE \"example.org%\")" {
		t.Errorf("Query is %s", query)
	}
}
