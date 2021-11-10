// TODO verify if any properties can be optional or have null as value.

export type NetworkFlow = {
    props: NetworkFlowProperties;
    lastSeenTimestamp: string; // ISO 8601 date string
};

export type NetworkFlowProperties = {
    srcEntity: NetworkEntityInfo;
    dstEntity: NetworkEntityInfo;
    dstPort: number; // uint32 may be 0 if not applicable (e.g., icmp)
    l4protocol: L4Protocol;
};

export type NetworkEndpoint = {
    props: NetworkEndpointProperties;
    lastActiveTimestamp: string; // ISO 8601 date string
};

export type NetworkEndpointProperties = {
    entity: NetworkEntityInfo;
    port: number; // uint32
    l4protocol: L4Protocol;
};

export type NetworkEntity = {
    info: NetworkEntityInfo;
    scope: NetworkEntityScope;
};

/*
 * Represents known cluster network peers to which the flows must be scoped.
 * In future, to restrict flows to more granular entities, such as deployment,
 * scope could include deployment ID.
 * Note: The highest scope level is cluster.
 */
export type NetworkEntityScope = {
    clusterId: string;
};

export type NetworkEntityInfo = DeploymentNetworkEntityInfo | ExternalSourceNetworkEntityInfo;

export type DeploymentNetworkEntityInfo = {
    deployment: {
        name: string;
        namespace: string;
        cluster: string; // deprecated
        listenPorts: ListenPort[];
    };
} & BaseNetworkEntityInfo;

export type ListenPort = {
    port: number; // uint32
    l4protocol: L4Protocol;
};

export type ExternalSourceNetworkEntityInfo = {
    externalSource: {
        name: string;
        cidr?: string;
        default: boolean; // `default` indicates whether the external source is user-generated or system-generated.
    };
} & BaseNetworkEntityInfo;

type BaseNetworkEntityInfo = {
    type: NetworkEntityInfoType;
    id: string;
};

export type NetworkEntityInfoType =
    | 'UNKNOWN_TYPE'
    | 'DEPLOYMENT'
    | 'INTERNET'
    | 'LISTEN_ENDPOINT'
    | 'EXTERNAL_SOURCE';

export type L4Protocol =
    | 'L4_PROTOCOL_UNKNOWN'
    | 'L4_PROTOCOL_TCP'
    | 'L4_PROTOCOL_UDP'
    | 'L4_PROTOCOL_ICMP'
    | 'L4_PROTOCOL_RAW'
    | 'L4_PROTOCOL_SCTP'
    | 'L4_PROTOCOL_ANY'; // -1