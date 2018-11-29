import { Injectable } from '@angular/core';
import { Http, Response} from '@angular/http';
import { Observable } from 'rxjs/Observable';
import "rxjs/add/observable/of";
import { HTTP_JSON_OPTIONS, buildHttpRequestOptions, HTTP_GET_OPTIONS } from "../shared/shared.utils";
import { RequestQueryParams } from 'harbor-ui';
import { DistributionHistory } from './distribution-history';
import { DistributionProvider, ProviderInstance, AuthMode } from './distribution-provider';


const providersEndpoint = '/api/distribution/providers';
const instanceEndpoint = '/api/distribution/instances';
const preheatEndpoint = '/api/distribution/preheats';

@Injectable()
export class DistributionService {

  constructor(private http: Http) { }

  private extractData(res: Response) {
    if (res.text() === '') {return []; };
    return res.json() || [];
  }

  private handleErrorObservable(error: Response | any) {
    console.error(error.message || error);
    return Observable.throw(error.message || error);
  }

  getDistributionHistories(params?: RequestQueryParams): Observable<DistributionHistory[]> {
    return this.http.get(preheatEndpoint, HTTP_GET_OPTIONS)
    .map(response => this.extractData(response))
    .catch(error => this.handleErrorObservable(error));
  }

  preheatImages(images: string[]): Observable<any> {
    return this.http
    .post(preheatEndpoint, {images: images},  HTTP_JSON_OPTIONS)
    .map(response => {
      return this.extractData(response);
    })
    .catch(this.handleErrorObservable);
  }

  getProviderInstances(params?: RequestQueryParams): Observable<ProviderInstance[]> {
    return this.http.get(instanceEndpoint, HTTP_GET_OPTIONS)
    .map(response => this.extractData(response))
    .catch(error => this.handleErrorObservable(error));
  }

  createProviderInstance(instance: any): Observable<any> {
    return this.http
    .post(instanceEndpoint, instance, HTTP_JSON_OPTIONS)
    .map(response => {
      return this.extractData(response);
    })
    .catch(this.handleErrorObservable);
  }

  updateProviderInstance(instanceID: string, instance: any): Observable<any> {
    return this.http
    .put(`${instanceEndpoint}/${instanceID}`, instance, HTTP_JSON_OPTIONS)
    .map(response => {
      return this.extractData(response);
    })
    .catch(this.handleErrorObservable);
  }

  deleteProviderInstance(instanceID: string): Observable<any> {
    return this.http
    .delete(`${instanceEndpoint}/${instanceID}`, HTTP_JSON_OPTIONS)
    .map(response => {
      return this.extractData(response);
    })
    .catch(this.handleErrorObservable);
  }

  getProviderDrivers(params?: RequestQueryParams): Observable<DistributionProvider[]> {
  return this.http.get(providersEndpoint, HTTP_GET_OPTIONS)
  .map(response => this.extractData(response))
  .catch(error => this.handleErrorObservable(error));
  }

}
