import {HttpClient} from "@angular/common/http";
import {Injectable} from "@angular/core";
import {Observable} from "rxjs";
import {Peer} from "../domain/peer";

@Injectable({
  providedIn: "root"
})
export class SeedService {

  constructor(private _httpClient: HttpClient) {
  }

  public getAllPeers(): Observable<Peer[]> {
    return this._httpClient.get<Peer[]>("/api/peers");
  }

  public postSeedList(list: string[]) {
    return this._httpClient.post<void>("/api/peers", list);
  }

}
