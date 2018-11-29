import { MessageHandlerService } from "./../../shared/message-handler/message-handler.service";
import { DistributionService } from "./../distribution.service";
import { Component, OnInit, ViewChild } from "@angular/core";
import {
  ProviderInstance,
  DistributionProvider,
  AuthMode
} from "../distribution-provider";
import { NgForm } from "@angular/forms";

@Component({
  selector: "dist-setup-modal",
  templateUrl: "./distribution-setup-modal.component.html",
  styleUrls: ["./distribution-setup-modal.component.scss"]
})
export class DistributionSetupModalComponent implements OnInit {
  opened: boolean = false;
  editingMode: boolean = false;
  model: ProviderInstance;
  basicUsername: string;
  basicPassword: string;
  authToken: string;
  @ViewChild("instanceForm") instanceForm: NgForm;

  constructor(
    private distributionService: DistributionService,
    private msgHandler: MessageHandlerService
  ) {}

  ngOnInit() {
    this.reset();
  }

  get title(): string {
    return this.editingMode ? "Edit Instance" : "Setup new instance";
  }

  _init() {
    //Init data
    let authData: Map<string, string> = new Map<string, string>();
    authData["username"] = "fake_user";
    authData["password"] = "fake_password";

    this.model = {
      id: this.uuid(),
      name: "",
      endpoint: "",
      status: "",
      enabled: true,
      setupTimestamp: new Date(),
      provider: "",
      auth_mode: AuthMode.BASIC,
      auth_data: authData
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
    let authData = {};
    if (this.model.auth_mode === AuthMode.BASIC) {
      authData[this.basicUsername] = this.basicPassword;
    } else if (this.model.auth_mode === AuthMode.OAUTH) {
      authData["token"] = this.authToken;
    }
    let instance = {};
    if (this.editingMode) {
      instance = {
        name: this.model.name,
        description: this.model.description,
        provider: "dragonfly",
        endpoint: this.model.endpoint,
        auth_mode: this.model.auth_mode,
        auth_data: authData,
        enabled: this.model.enabled
      };
      this.distributionService
        .updateProviderInstance(this.model.id, instance)
        .subscribe(() => this.msgHandler.info, () => this.msgHandler.error);
    } else {
      instance = {
        name: this.model.name,
        description: this.model.description,
        provider: "dragonfly",
        endpoint: this.model.endpoint,
        auth_mode: this.model.auth_mode,
        auth_data: authData,
        enabled: this.model.enabled
      };
      this.distributionService
        .createProviderInstance(instance)
        .subscribe(() => this.msgHandler.info, () => this.msgHandler.error);
    }

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
    return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, char => {
      let random = (Math.random() * 16) | 0;
      let value = char === "x" ? random : (random % 4) + 8;
      return value.toString(16);
    });
  }
}
