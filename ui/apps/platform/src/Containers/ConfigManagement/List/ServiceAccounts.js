import React from 'react';
import pluralize from 'pluralize';

import {
    defaultHeaderClassName,
    defaultColumnClassName,
    nonSortableHeaderClassName,
} from 'Components/Table';
import { entityListPropTypes, entityListDefaultprops } from 'constants/entityPageProps';
import entityTypes from 'constants/entityTypes';
import { serviceAccountSortFields } from 'constants/sortFields';
import { SERVICE_ACCOUNTS_QUERY } from 'queries/serviceAccount';
import { sortValueByLength } from 'sorters/sorters';
import queryService from 'utils/queryService';
import URLService from 'utils/URLService';
import List from './List';
import TableCellLink from './Link';

export const defaultServiceAccountSort = [
    {
        id: serviceAccountSortFields.SERVCE_ACCOUNT,
        desc: false,
    },
];
const buildTableColumns = (match, location, entityContext) => {
    const tableColumns = [
        {
            Header: 'Id',
            headerClassName: 'hidden',
            className: 'hidden',
            accessor: 'id',
        },
        {
            Header: `Service Accounts`,
            headerClassName: `w-1/10 ${defaultHeaderClassName}`,
            className: `w-1/10 ${defaultColumnClassName}`,
            accessor: 'name',
            id: serviceAccountSortFields.SERVCE_ACCOUNT,
            sortField: serviceAccountSortFields.SERVCE_ACCOUNT,
        },
        {
            Header: `Cluster Admin Role`,
            headerClassName: `w-1/10 ${nonSortableHeaderClassName}`,
            className: `w-1/10 ${defaultColumnClassName}`,
            Cell: ({ original }) => {
                const { clusterAdmin } = original;
                return clusterAdmin ? 'Enabled' : 'Disabled';
            },
            accessor: 'clusterAdmin',
            sortable: false,
        },
        entityContext && entityContext[entityTypes.CLUSTER]
            ? null
            : {
                  Header: `Cluster`,
                  headerClassName: `w-1/8 ${defaultHeaderClassName}`,
                  className: `w-1/8 ${defaultColumnClassName}`,
                  accessor: 'clusterName',
                  // eslint-disable-next-line
                  Cell: ({ original, pdf }) => {
                      const { clusterName, clusterId, id } = original;
                      const url = URLService.getURL(match, location)
                          .push(id)
                          .push(entityTypes.CLUSTER, clusterId)
                          .url();
                      return <TableCellLink pdf={pdf} url={url} text={clusterName} />;
                  },
                  id: serviceAccountSortFields.CLUSTER,
                  sortField: serviceAccountSortFields.CLUSTER,
              },
        entityContext && entityContext[entityTypes.NAMESPACE]
            ? null
            : {
                  Header: `Namespace`,
                  headerClassName: `w-1/10 ${defaultHeaderClassName}`,
                  className: `w-1/10 ${defaultColumnClassName}`,
                  accessor: 'namespace',
                  // eslint-disable-next-line
                  Cell: ({ original, pdf }) => {
                      const {
                          id,
                          saNamespace: { metadata },
                      } = original;
                      if (!metadata) {
                          return 'No Matches';
                      }
                      const { name, id: namespaceId } = metadata;
                      const url = URLService.getURL(match, location)
                          .push(id)
                          .push(entityTypes.NAMESPACE, namespaceId)
                          .url();
                      return <TableCellLink pdf={pdf} url={url} text={name} />;
                  },
                  id: serviceAccountSortFields.NAMESPACE,
                  sortField: serviceAccountSortFields.NAMESPACE,
              },
        {
            Header: `Roles`,
            headerClassName: `w-1/8 ${nonSortableHeaderClassName}`,
            className: `w-1/8 ${defaultColumnClassName}`,
            Cell: ({ original, pdf }) => {
                const { id, k8sRoles } = original;
                const { length } = k8sRoles;
                if (!length) {
                    return 'No Roles';
                }
                const url = URLService.getURL(match, location)
                    .push(id)
                    .push(entityTypes.ROLE)
                    .url();
                if (length > 1) {
                    return (
                        <TableCellLink
                            pdf={pdf}
                            url={url}
                            text={`${length} ${pluralize('Roles', length)}`}
                        />
                    );
                }
                return original.k8sRoles[0].name;
            },
            accessor: 'k8sRoles',
            sortMethod: sortValueByLength,
            sortable: false,
        },
        {
            Header: `Deployments`,
            headerClassName: `w-1/8 ${nonSortableHeaderClassName}`,
            className: `w-1/8 ${defaultColumnClassName}`,
            // eslint-disable-next-line
            Cell: ({ original, pdf }) => {
                const { id, deploymentCount } = original;
                if (!deploymentCount) {
                    return 'No Deployments';
                }
                const url = URLService.getURL(match, location)
                    .push(id)
                    .push(entityTypes.DEPLOYMENT)
                    .url();
                return (
                    <TableCellLink
                        pdf={pdf}
                        url={url}
                        text={`${deploymentCount} ${pluralize('Deployment', deploymentCount)}`}
                    />
                );
            },
            accessor: 'deploymentCount',
            sortable: false,
        },
    ];
    return tableColumns.filter((col) => col);
};

const createTableRows = (data) => data.results;

const ServiceAccounts = ({
    match,
    location,
    className,
    selectedRowId,
    onRowClick,
    query,
    data,
    totalResults,
    entityContext,
}) => {
    const autoFocusSearchInput = !selectedRowId;
    const tableColumns = buildTableColumns(match, location, entityContext);
    const queryText = queryService.objectToWhereClause(query);
    const variables = queryText ? { query: queryText } : null;
    return (
        <List
            className={className}
            query={SERVICE_ACCOUNTS_QUERY}
            variables={variables}
            entityType={entityTypes.SERVICE_ACCOUNT}
            tableColumns={tableColumns}
            createTableRows={createTableRows}
            onRowClick={onRowClick}
            selectedRowId={selectedRowId}
            idAttribute="id"
            defaultSorted={defaultServiceAccountSort}
            data={data}
            totalResults={totalResults}
            autoFocusSearchInput={autoFocusSearchInput}
        />
    );
};
ServiceAccounts.propTypes = entityListPropTypes;
ServiceAccounts.defaultProps = entityListDefaultprops;

export default ServiceAccounts;
