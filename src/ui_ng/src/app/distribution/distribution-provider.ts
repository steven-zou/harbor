export interface DistributionProvider {
    name: string;
    icon: string;
    version: string;
    source: string;
    maintainers: string[];
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
}
