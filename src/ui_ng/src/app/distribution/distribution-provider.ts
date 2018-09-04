export interface DistributionProvider {
    name: string;
    icon: string;
    version: string;
    source: string;
    maintainers: string[];
    authMode: string;
}

export interface ProviderInstance {
    ID: string;
    name: string;
    endpoint: string;
    description?: string;
    status: string;
    enabled: boolean;
    setupTimestamp: Date;
    provider: DistributionProvider;
    authorization: AuthorizationData;
}

export class AuthMode {
    static BASIC = "BASIC";
    static OAUTH = "OAUTH";
    static CUSTOM = "CUSTOM";
}

export interface AuthorizationData {
    authMode: string;
    //Keep the auth data
    //if authMode is 'BASIC', then 'username' and 'password' are stored;
    //if authMode is 'OAUTH', then 'token' is stored'
    //if authMode is 'CUSTOM', then 'header_key' with corresponding header value are stored.
    data: Map<string, string>;
}
