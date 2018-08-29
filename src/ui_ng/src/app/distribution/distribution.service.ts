import { Injectable } from '@angular/core';
import { Http } from '@angular/http';
import { Observable } from 'rxjs/Observable';
import "rxjs/add/observable/of";
import { HTTP_JSON_OPTIONS, buildHttpRequestOptions } from "../shared/shared.utils";
import { RequestQueryParams } from 'harbor-ui';
import { DistributionHistory } from './distribution-history';

@Injectable()
export class DistributionService {

  constructor(private http: Http) { }

  getDistributionHistories(params?: RequestQueryParams):Observable<DistributionHistory[]>{
    let mockData: DistributionHistory[] = [
      {
        image: "library/redis:latest",
        timestamp: new Date(),
        status: "PENDING",
        provider: "Dragonfly",
        instance: "uuid"
      }
    ];
    return Observable.of(mockData);
  }

}
