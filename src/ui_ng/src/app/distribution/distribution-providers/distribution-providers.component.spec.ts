import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { DistributionProvidersComponent } from './distribution-providers.component';

describe('DistributionProvidersComponent', () => {
  let component: DistributionProvidersComponent;
  let fixture: ComponentFixture<DistributionProvidersComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ DistributionProvidersComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(DistributionProvidersComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
