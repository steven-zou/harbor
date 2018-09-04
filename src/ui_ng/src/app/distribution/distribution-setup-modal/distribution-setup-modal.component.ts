import { Component, OnInit, ViewChild } from '@angular/core';
import { ProviderInstance, DistributionProvider, AuthMode } from '../distribution-provider';
import { NgForm } from '@angular/forms';

@Component({
  selector: 'dist-setup-modal',
  templateUrl: './distribution-setup-modal.component.html',
  styleUrls: ['./distribution-setup-modal.component.scss']
})
export class DistributionSetupModalComponent implements OnInit {

  opened: boolean = false;
  editingMode: boolean = false;
  model: ProviderInstance;
  @ViewChild("instanceForm") instanceForm: NgForm;

  constructor() { }

  ngOnInit() {
    this.reset();
  }

  get title(): string {
    return this.editingMode ? "Edit Instance" : "Setup new instance"
  }

  get authMode(): string {
    return this.model && this.model.provider ?
      this.model.provider.authMode : AuthMode.BASIC;
  }

  _init() {
    //Init data
    let authData: Map<string, string> = new Map<string, string>();
    authData["username"] = "fake_user";
    authData["password"] = "fake_password";

    this.model = {
      ID: this.uuid(),
      name: "",
      endpoint: "",
      status: "",
      enabled: true,
      setupTimestamp: new Date(),
      provider: null,
      authorization: {
        authMode: AuthMode.BASIC,
        data: authData
      }
    };
  }

  _open() {
    this.opened = true;
  }

  _close() {
    this.opened = false;
  }

  _isProvider(obj: any): obj is DistributionProvider {
    return obj.version !== undefined;
  }

  _isInstance(obj: any): obj is ProviderInstance {
    return obj.endpoint !== undefined;
  }

  get enabledLabel(): string {
    return this.model && this.model.enabled ? "ON" : "OFF";
  }

  cancel() {
    this._close();
  }

  submit() {
    console.log(this.model);
    this._close();
  }

  openSetupModal(mode: string, data: any) {
    this.editingMode = mode === "edit" ? true : false;
    this._open();

    if (this._isProvider(data)) {
      this.model.provider = data;
      return;
    }

    if (this._isInstance(data)) {
      this.model = data;
      return;
    }
  }

  reset() {
    this._init();
    this.instanceForm.reset();
  }

  uuid(): string {
    return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (char) => {
      let random = Math.random() * 16 | 0;
      let value = char === "x" ? random : (random % 4 + 8);
      return value.toString(16);
    });
  }
}
