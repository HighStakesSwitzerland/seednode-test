import {HttpErrorResponse} from "@angular/common/http";
import {Component, Inject} from "@angular/core";
import {MAT_DIALOG_DATA} from "@angular/material/dialog";
import {isArray as _isArray} from "lodash-es"
@Component({
  selector: "app-dialog",
  templateUrl: "./dialog.component.html",
  styleUrls: ["./dialog.component.css"]
})
export class DialogComponent {

  public _isArray = _isArray;

  constructor(@Inject(MAT_DIALOG_DATA) public data: DialogData) {
  }

}

export interface DialogData {
  title: string;
  errors: string[];
  httpError: HttpErrorResponse;
}
