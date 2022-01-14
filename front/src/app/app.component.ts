import {Component, ViewChild} from "@angular/core";
import {MatTabGroup} from "@angular/material/tabs";

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {

  @ViewChild(MatTabGroup)
  public matTabGroup! : MatTabGroup;
}
