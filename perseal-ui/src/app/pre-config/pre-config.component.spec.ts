import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { PreConfigComponent } from './pre-config.component';

describe('PerLoadComponent', () => {
  let component: PreConfigComponent;
  let fixture: ComponentFixture<PreConfigComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ PreConfigComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(PreConfigComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
