import React, { Component } from 'react';
import PropTypes from 'prop-types';
import { connect } from 'react-redux';
import { createStructuredSelector, createSelector } from 'reselect';
import { selectors } from 'reducers';
import { formValues } from 'redux-form';
import uniq from 'lodash/uniq';
import uniqBy from 'lodash/uniqBy';

import { components } from 'react-select';
import * as Icon from 'react-feather';

import Labeled from 'Components/Labeled';
import ReduxSelectField from 'Components/forms/ReduxSelectField';
import ReduxSelectCreatableField from 'Components/forms/ReduxSelectCreatableField';
import { selectMenuOnTopStyles } from 'Components/ReactSelect';

const standardKeyOptions = ['userid', 'name', 'email', 'groups'];
const defaultKeyOptionsByType = {
    userpki: ['CN', 'C', 'O', 'OU', 'L', 'ST', 'STREET', 'POSTALCODE', 'DN'],
};

const MenuList = ({ toggleModal, ...props }) => (
    <components.MenuList {...props}>
        {props.children}
        <div className="border-b border-t border-primary-400 w-full p-3">
            <button type="button" className="btn btn-primary w-full" onClick={toggleModal}>
                Create New Role
            </button>
        </div>
    </components.MenuList>
);

class RuleGroups extends Component {
    static propTypes = {
        initialValues: PropTypes.shape({
            id: PropTypes.string,
            type: PropTypes.string,
            groups: PropTypes.arrayOf(
                PropTypes.shape({
                    props: PropTypes.shape({
                        key: PropTypes.string,
                    }),
                })
            ),
        }),
        groups: PropTypes.arrayOf(
            PropTypes.shape({
                props: PropTypes.shape({
                    key: PropTypes.string,
                }),
            })
        ).isRequired,
        fields: PropTypes.arrayOf(PropTypes.shape({})).isRequired,
        toggleModal: PropTypes.func.isRequired,
        roles: PropTypes.arrayOf(
            PropTypes.shape({
                name: PropTypes.string,
                globalAccess: PropTypes.string,
            })
        ).isRequired,
        usersAttributes: PropTypes.arrayOf(
            PropTypes.shape({
                authProviderId: PropTypes.string,
                key: PropTypes.string,
                value: PropTypes.string,
            })
        ).isRequired,
    };

    static defaultProps = {
        initialValues: {
            id: '',
        },
    };

    renderMenuListComponent = (props) => (
        <MenuList toggleModal={this.props.toggleModal} {...props} />
    );

    getFilteredValueOptions = (valueOptions, idx) => {
        const { key } = this.props?.initialValues?.groups[idx]?.props ?? {};
        const result = valueOptions
            .filter((option) => option.key === key)
            .map((option) => ({ label: option.label, value: option.value }));
        return result;
    };

    render() {
        const { fields, initialValues, usersAttributes, roles } = this.props;
        const defaultKeyOptions = defaultKeyOptionsByType[initialValues.type] || standardKeyOptions;
        const keyOptions = uniq([
            ...defaultKeyOptions,
            ...usersAttributes
                .filter(({ authProviderId }) => authProviderId === initialValues.id)
                .map(({ key }) => key),
        ]).map((v) => ({ value: v, label: v }));

        let valueOptions = initialValues.groups.map(({ props: { key, value } }) => ({
            key,
            label: value,
            value,
        }));
        valueOptions = uniqBy(
            usersAttributes
                .map(({ key, value }) => ({
                    key,
                    label: value,
                    value,
                }))
                .concat(valueOptions),
            'value'
        );
        const addRule = () => fields.push({ props: { auth_provider_id: initialValues.id } });
        const deleteRule = (group, idx) => () => {
            fields.remove(idx);
        };
        return (
            <div className="w-full p-2">
                {fields.map((group, idx, value) => (
                    <div className="flex" key={idx}>
                        <div className="w-full">
                            <Labeled label="Key">
                                <ReduxSelectCreatableField
                                    name={`${group}.props.key`}
                                    options={keyOptions}
                                    styles={selectMenuOnTopStyles}
                                />
                            </Labeled>
                        </div>
                        <div className="w-full">
                            <Labeled label="Value">
                                <ReduxSelectCreatableField
                                    name={`${group}.props.value`}
                                    options={this.getFilteredValueOptions(valueOptions, idx)}
                                    styles={selectMenuOnTopStyles}
                                />
                            </Labeled>
                        </div>
                        <div className="flex items-center mt-2">
                            <Icon.ArrowRight className="h-4 w-4" />
                        </div>
                        <div className="w-full">
                            <Labeled label="Role">
                                <ReduxSelectField
                                    name={`${group}.roleName`}
                                    options={roles}
                                    customComponents={{
                                        MenuList: this.renderMenuListComponent,
                                    }}
                                    styles={selectMenuOnTopStyles}
                                />
                            </Labeled>
                        </div>
                        <button className="pl-2 pr-2 mt-2" type="button">
                            <Icon.Trash2 className="h-4 w-4" onClick={deleteRule(value, idx)} />
                        </button>
                    </div>
                ))}
                {/* eslint-disable */}
                <button
                    className="border-2 bg-primary-200 border-primary-400 text-sm text-primary-700 hover:bg-primary-300 hover:border-primary-500 rounded-sm block px-3 py-2 uppercase ml-1 mb-4"
                    type="button"
                    onClick={addRule}
                >
                    {/* eslint-enable */}
                    Add New Rule
                </button>
            </div>
        );
    }
}

const getRoleOptions = createSelector([selectors.getRoles], (roles) =>
    roles.map((role) => ({ value: role.name, label: role.name }))
);

const mapStateToProps = createStructuredSelector({
    usersAttributes: selectors.getUsersAttributes,
    roles: getRoleOptions,
});

const mapDispatchToProps = {};

export default connect(mapStateToProps, mapDispatchToProps)(formValues('groups')(RuleGroups));