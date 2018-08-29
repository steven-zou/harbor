import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { DistributionHistoryComponent } from './distribution-history.component';

describe('DistributionHistoryComponent', () => {
  let component: DistributionHistoryComponent;
  let fixture: ComponentFixture<DistributionHistoryComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ DistributionHistoryComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(DistributionHistoryComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
