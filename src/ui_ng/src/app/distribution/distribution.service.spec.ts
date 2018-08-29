import { TestBed, inject } from '@angular/core/testing';

import { DistributionService } from './distribution.service';

describe('DistributionService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [DistributionService]
    });
  });

  it('should be created', inject([DistributionService], (service: DistributionService) => {
    expect(service).toBeTruthy();
  }));
});
