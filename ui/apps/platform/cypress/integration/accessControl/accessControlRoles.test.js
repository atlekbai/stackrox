import { rolesUrl, selectors } from '../../constants/AccessControlPage';
import { permissions as permissionsApi } from '../../constants/apiEndpoints';

import withAuth from '../../helpers/basicAuth';
import { hasFeatureFlag } from '../../helpers/features';

// Migration from cy.server and cy.route to cy.intercept fails for /v1/roles/* imported from apiEndpoints.
const rolesApi = {
    list: '/v1/roles',
};

const h1 = 'Access Control';
const h2 = 'Roles';

const defaultNames = ['Admin', 'Analyst', 'Continuous Integration', 'None', 'Sensor Creator'];

describe('Access Control Roles', () => {
    withAuth();

    before(function beforeHook() {
        if (!hasFeatureFlag('ROX_SCOPED_ACCESS_CONTROL_V2')) {
            this.skip();
        }
    });

    function visitRoles() {
        cy.intercept('GET', rolesApi.list).as('GetRoles');
        cy.visit(rolesUrl);
        cy.wait('@GetRoles');
    }

    function visitRole(name) {
        cy.intercept('GET', rolesApi.list).as('GetRoles');
        cy.visit(`${rolesUrl}/${name}`);
        cy.wait('@GetRoles');
    }

    it('displays alert if no permission', () => {
        cy.intercept('GET', permissionsApi.mypermissions, {
            fixture: 'auth/mypermissionsMinimalAccess.json',
        }).as('GetMyPermissions');
        cy.visit(rolesUrl);
        cy.wait('@GetMyPermissions');

        cy.get(selectors.h1).should('have.text', h1);
        cy.get(selectors.navLink).should('not.exist');

        cy.get(selectors.h2).should('not.exist');

        cy.get(selectors.alertTitle).should(
            'contain', // not have.text because it contains "Info alert:" for screen reader
            'You do not have permission to view Access Control'
        );
    });

    it('list has breadcrumbs, headings, link, button, and table head cells', () => {
        visitRoles();

        cy.get(`${selectors.breadcrumbItem}:nth-child(1):contains("${h1}")`);
        cy.get(`${selectors.breadcrumbItem}:nth-child(2):contains("${h2}")`);

        cy.get(selectors.h1).should('have.text', h1);
        cy.get(selectors.navLinkCurrent).should('have.text', h2);

        cy.get(selectors.h2).should('have.text', h2);
        cy.get(selectors.list.addButton).should('have.text', 'Add role');

        cy.get(`${selectors.list.th}:contains("Name")`);
        cy.get(`${selectors.list.th}:contains("Description")`);
        cy.get(`${selectors.list.th}:contains("Permission set")`);
        cy.get(`${selectors.list.th}:contains("Access scope")`);
    });

    it('list has default names', () => {
        visitRoles();

        const { tdPermissionSetLink, tdAccessScope } = selectors.list.roles;

        cy.get(selectors.list.tdNameLink).then(($tds) => {
            $tds.get().forEach((td, index) => {
                const roleName = td.textContent;
                if (defaultNames.includes(roleName)) {
                    cy.get(`${tdPermissionSetLink}:eq(${index})`).should('have.text', roleName);
                    cy.get(`${tdAccessScope}:eq(${index})`).should('have.text', 'Unrestricted');
                }
            });
        });
    });

    it('list link goes to form which has label instead of button and disabled input values', () => {
        visitRoles();

        const name = 'Admin';
        cy.get(`${selectors.list.tdNameLink}:contains("${name}")`).click();

        cy.get(selectors.h1).should('have.text', h1);
        cy.get(selectors.navLinkCurrent).should('have.text', h2);

        cy.get(selectors.h2).should('have.text', name);
        cy.get(selectors.form.notEditableLabel).should('exist');
        cy.get(selectors.form.editButton).should('not.exist');

        cy.get(selectors.form.inputName).should('be.disabled');
        cy.get(selectors.form.inputDescription).should('be.disabled');

        const { getRadioPermissionSetForName, getRadioAccessScopeForName } = selectors.form.role;

        defaultNames.forEach((defaultName) => {
            cy.get(getRadioPermissionSetForName(defaultName)).should('be.disabled');
        });

        cy.get(getRadioAccessScopeForName('Unrestricted')).should('be.disabled');
    });

    describe('direct link to default role', () => {
        const { getRadioPermissionSetForName, getRadioAccessScopeForName } = selectors.form.role;

        defaultNames.forEach((roleName) => {
            it(`${roleName} has corresponding permission set and no access scope`, () => {
                visitRole(roleName);

                cy.get(selectors.form.inputName).should('have.value', roleName);

                defaultNames.forEach((defaultName) => {
                    cy.get(getRadioPermissionSetForName(defaultName)).should(
                        defaultName === roleName ? 'be.checked' : 'not.be.checked'
                    );
                });

                cy.get(getRadioAccessScopeForName('Unrestricted')).should('be.checked');
            });
        });
    });

    it('creates a new role and form disables name input when editing an existing role', () => {
        visitRoles();

        cy.get(selectors.list.addButton).click();

        cy.get(selectors.h2).should('have.text', 'Add role');
        cy.get(selectors.form.notEditableLabel).should('not.exist');
        cy.get(selectors.form.editButton).should('not.exist');
        cy.get(selectors.form.saveButton).should('be.disabled');
        cy.get(selectors.form.cancelButton).should('be.enabled');

        cy.get(selectors.form.inputName).should('be.enabled').should('have.value', '');
        cy.get(selectors.form.inputDescription).should('be.enabled').should('have.value', '');

        const name = `Role-${new Date().getTime()}`;
        const description = 'New description';
        const permissionSetName = 'None';
        const accessScopeName = 'Unrestricted';

        cy.get(selectors.form.inputName).type(name);
        cy.get(selectors.form.inputDescription).type(description);
        cy.get(selectors.form.role.getRadioPermissionSetForName(permissionSetName)).click();
        cy.get(selectors.form.role.getRadioAccessScopeForName(accessScopeName))
            .should('be.enabled')
            .should('be.checked');

        cy.intercept('POST', rolesApi.list).as('PostRoles');
        cy.get(selectors.form.saveButton).click();
        cy.wait('@PostRoles');

        cy.get(selectors.h2).should('have.text', name);
        cy.get(selectors.form.inputName).should('be.disabled').should('have.value', name);
        cy.get(selectors.form.notEditableLabel).should('not.exist');
        cy.get(selectors.form.editButton).should('exist');
        cy.get(selectors.form.saveButton).should('not.exist');
        cy.get(selectors.form.cancelButton).should('not.exist');

        cy.get(selectors.form.editButton).click();
        cy.get(selectors.form.editButton).should('be.disabled');
        cy.get(selectors.form.inputName).should('be.disabled');

        cy.get(selectors.form.cancelButton).click();
        cy.get(selectors.form.editButton).should('be.enabled');
    });
});