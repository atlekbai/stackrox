import randomstring from 'randomstring';

import { selectors, url } from '../../constants/ViolationsPage';
import search from '../../selectors/search';
import * as api from '../../constants/apiEndpoints';
import withAuth from '../../helpers/basicAuth';

function setAlertRoutes() {
    cy.server();
    cy.route('GET', api.alerts.alerts).as('alerts');
    cy.route('GET', api.alerts.alertById).as('alertById');
    cy.route('POST', api.graphql(api.alerts.graphqlOps.getTags)).as('getTags');
    cy.route('POST', api.graphql(api.alerts.graphqlOps.tagsAutocomplete)).as('tagsAutocomplete');
    cy.route('POST', api.graphql(api.alerts.graphqlOps.bulkAddAlertTags)).as('bulkAddAlertTags');
}

function openFirstItemOnViolationsPage() {
    cy.visit(url);
    cy.wait('@alerts');

    cy.get(selectors.firstTableRowLink).click();
    cy.wait('@alertById');
    cy.wait(['@getTags', '@tagsAutocomplete']);
}

function enterPageSearch(searchObj, inputSelector = search.input) {
    cy.route(api.search.autocomplete).as('searchAutocomplete');
    function selectSearchOption(optionText) {
        // typing is slow, assuming we'll get autocomplete results, select them
        // also, likely it'll mimic better typical user's behavior
        cy.get(inputSelector).type(`${optionText.charAt(0)}`);
        cy.wait('@searchAutocomplete');
        cy.get(search.options).contains(optionText).first().click({ force: true });
    }

    Object.entries(searchObj).forEach(([searchCategory, searchValue]) => {
        selectSearchOption(searchCategory);

        if (Array.isArray(searchValue)) {
            searchValue.forEach((val) => selectSearchOption(val));
        } else {
            selectSearchOption(searchValue);
        }
    });
    cy.get(inputSelector).blur(); // remove focus to close the autocomplete popup
}

describe('Violation Page: Tags', () => {
    withAuth();

    it('should add tag without allowing duplicates', () => {
        setAlertRoutes();
        openFirstItemOnViolationsPage();

        const tag = randomstring.generate(7);
        cy.get(selectors.details.tags.input).type(`${tag}{enter}`);
        // do it again to check that no duplicate tags can be added
        cy.get(selectors.details.tags.input).type(`${tag}{enter}`);
        cy.wait(['@getTags', '@tagsAutocomplete']);

        // pressing {enter} won't save the tag, only one would be displayed as tag chip
        cy.get(selectors.details.tags.values).contains(tag).should('have.length', 1);
    });

    it('should add tag without allowing duplicates with leading/trailing whitespace', () => {
        setAlertRoutes();
        openFirstItemOnViolationsPage();

        const tag = randomstring.generate(7);
        cy.get(selectors.details.tags.input).type(`${tag}{enter}`);
        // do it again to check that no duplicate tags can be added
        cy.get(selectors.details.tags.input).type(`   ${tag}   {enter}`);
        cy.wait(['@getTags', '@tagsAutocomplete']);

        // pressing {enter} won't save the tag, only one would be displayed as tag chip
        cy.get(selectors.details.tags.values).contains(tag).should('have.length', 1);
    });

    it('should add bulk tags without duplication', () => {
        setAlertRoutes();

        cy.visit(url);
        cy.wait('@alerts');

        // check first item
        cy.get(`${selectors.firstTableRow} input[type="checkbox"]`).should('not.be.checked');
        cy.get(`${selectors.firstTableRow} input[type="checkbox"]`).check();

        // add tags
        cy.get(selectors.actions.dropdown).click();
        cy.get(selectors.actions.addTagsBtn).click();
        const tag = randomstring.generate(7);
        cy.get(selectors.modal.tagConfirmation.input).type(`${tag}{enter}`);
        cy.get(selectors.modal.tagConfirmation.confirmBtn).click();
        cy.wait('@bulkAddAlertTags');
        cy.wait(1000);

        cy.get(`${selectors.firstTableRow} input[type="checkbox"]`).should('not.be.checked');
        cy.get(`${selectors.firstTableRow} input[type="checkbox"]`).check();
        // also check some other violation
        cy.get(
            `${selectors.table.rows}:not(${selectors.firstTableRow}):first input[type="checkbox"]`
        ).should('not.be.checked');
        cy.get(
            `${selectors.table.rows}:not(${selectors.firstTableRow}):first input[type="checkbox"]`
        ).check();

        cy.get(selectors.actions.dropdown).click();
        cy.get(selectors.actions.addTagsBtn).click();
        // ROX-4626: until we hit {enter} the tag isn't created yet, button should be disabled
        cy.get(selectors.modal.tagConfirmation.confirmBtn).should('be.disabled');

        cy.get(selectors.modal.tagConfirmation.input).type(`${tag}{enter}`);
        cy.get(selectors.modal.tagConfirmation.confirmBtn).click();
        cy.wait('@bulkAddAlertTags');

        enterPageSearch({ Tag: tag });
        cy.wait('@alerts');

        cy.get(selectors.table.rows).should('have.length', 2);
    });

    it('should suggest autocompletion for existing tags', () => {
        setAlertRoutes();
        openFirstItemOnViolationsPage();

        const tag = randomstring.generate(7);
        cy.get(selectors.details.tags.input).type(`${tag}{enter}`);
        cy.wait(['@getTags', '@tagsAutocomplete']);

        cy.visit(url);
        cy.wait('@alerts');

        // check bulk dialog autocompletion
        cy.get(`${selectors.firstTableRow} input[type="checkbox"]`)
            .should('not.be.checked')
            .check();
        cy.get(selectors.actions.dropdown).click();
        cy.get(selectors.actions.addTagsBtn).click();
        cy.get(selectors.modal.tagConfirmation.input).type(`${tag.charAt(0)}`);
        cy.get(`${selectors.modal.tagConfirmation.options}:contains("${tag}")`).should('exist');
    });

    it('should remove tag', () => {
        setAlertRoutes();
        openFirstItemOnViolationsPage();

        const tag = randomstring.generate(7);
        cy.get(selectors.details.tags.input).type(`${tag}{enter}`);
        cy.wait(['@getTags', '@tagsAutocomplete']);

        cy.get(selectors.details.tags.removeValueButton(tag)).click();
        cy.wait(['@getTags', '@tagsAutocomplete']);

        cy.get(`${selectors.details.tags.values}:contains("${tag}")`).should('not.exist');
    });
});
