export interface DistributionProvider {
    name: string;
    icon: string;
    version: string;
    source: string;
    maintainers: string[];
    auth_mode: string;
}

export interface ProviderInstance {
    id: string;
    name: string;
    endpoint: string;
    description?: string;
    status: string;
    enabled: boolean;
    setupTimestamp?: Date;
    provider: DistributionProvider | string;
    auth_mode: string;
    auth_data: any;
}

export class AuthMode {
    static NONE = "NONE";
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
