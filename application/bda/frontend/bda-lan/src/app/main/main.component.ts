import { Component, OnInit, OnDestroy, ViewChild } from '@angular/core';
import { FormControl, Validators, NgForm } from '@angular/forms';
import { HttpClient } from '@angular/common/http'
import { AuthService } from '../services/auth.service'
import { Router } from '@angular/router';
import { MatSnackBar } from '@angular/material';
import { MatPaginator, MatTableDataSource } from '@angular/material';
import { Subscription } from 'rxjs';

export interface viewLayout {
  assetType: string,
  id: string,
  address: string,
  noc: string,
  la: string,
  fa: string,
  tot: string
}

export interface layoutHistory {  
  TxId: string,
  Id: string,
  Address: string,
  requestedNOC: boolean,
  FAStatus: string,
  LAStatus: string,
  ApprovalStatus: string,
  TimeStamp: string,
  IsDelete: boolean
}


@Component({
  selector: 'app-main',
  templateUrl: './main.component.html',
  styleUrls: ['./main.component.scss']
})
export class MainComponent implements OnInit, OnDestroy {

  host = "10.199.1.143"; // Change in auth service also
  id = new FormControl('', [Validators.required]);
  address = new FormControl('', [Validators.required]);
  viewLayoutResult: {key: string, value: string}[] = [];
  layoutHistoryResult: {key: string, value: string}[][] = [];
  isLoading = false; 
  dataSource = new MatTableDataSource<layoutHistory>(); 
  displayedColumns: string[];
  historyLoaded = false;

  @ViewChild(MatPaginator) paginator: MatPaginator;
  userIsAuthenticated = false;
  private authListenerSubs: Subscription;

  constructor(private http: HttpClient,
              private authService: AuthService,
              private router: Router,
              private snackBar: MatSnackBar) { }

  ngOnInit() {    
    this.userIsAuthenticated = this.authService.getIsAuth();
    this.authListenerSubs = this.authService
      .getAuthStatusListener()
      .subscribe(isAuthenticated => {
        this.userIsAuthenticated = isAuthenticated;
      });
    if(!this.userIsAuthenticated) {
      this.router.navigate(["/login"]);
    }
    this.dataSource.paginator = this.paginator;
  }  

  ngOnDestroy() {
    this.authListenerSubs.unsubscribe();
  }

  createLayout(form: NgForm){
    if(form.invalid) {
      return;
    }
    const token = this.authService.getToken();
    const obj = {id: form.value.id, address: form.value.address, token: token};
    this.isLoading = true;
    this.http
      .post<{ message: string }>(
        `http://${this.host}:3000/api/createlayout`, obj
      )
      .subscribe(response => {
          this.isLoading = false;
          const message = response.message;
          console.log("Response message: " + message);
          this.snackBar.open(message, "OK", {
            duration: 0,
          });

      });
    form.resetForm(''); 
  }

  viewLayout (form: NgForm) {
    if(form.invalid) {
      return;
    }
    const token = this.authService.getToken();
    const obj = {id: form.value.id, token: token};
    this.isLoading = true;
    this.http
      .post<{ status: number,message: string, result: string }>(
        `http://${this.host}:3000/api/viewlayout`, obj
      )
      .subscribe(response => {
          this.isLoading = false;
          const message = response.message;
          const status = response.status;
          this.viewLayoutResult = [];
          if(status == 200) {
            const result = JSON.parse(response.result);
            this.viewLayoutResult.push({key: "Id", value: result.Id}); 
            this.viewLayoutResult.push({key: "Address", value: result.Address}); 
            this.viewLayoutResult.push({key: "RequestedNOC", value: result.requestedNOC}); 
            this.viewLayoutResult.push({key: "LAStatus", value: result.LAStatus}); 
            this.viewLayoutResult.push({key: "FAStatus", value: result.FAStatus}); 
            this.viewLayoutResult.push({key: "ApprovalStatus", value: result.ApprovalStatus});
          }  
          console.log("Response message: " + message);
          this.snackBar.open(message, "OK", {
            duration: 0,
          });

      });
    form.resetForm(''); 

  }

  requestNOC(form: NgForm) { 
    if(form.invalid) {
      return;
    }  
    const token = this.authService.getToken();
    const obj = {id: form.value.id, token: token};
    this.isLoading = true;
    this.http
      .post<{ message: string }>(
        `http://${this.host}:3000/api/requestNOC`, obj
      )
      .subscribe(response => {
          this.isLoading = false;
          const message = response.message;
          console.log("Response message: " + message);
          this.snackBar.open(message, "OK", {
            duration: 0,
          });

      });
    form.resetForm(''); 
    
  }

  getHistory (form: NgForm) {
    console.log("ID: "+form.value.id);
    var id = form.value.id;
    if(form.value.id == null){
      id = "ALL_TRANSACTION_HISTORY";
    }
    if(form.invalid) {
      return;
    }
    const token = this.authService.getToken();
    const obj = {id: id, token: token};
    this.isLoading = true;
    this.historyLoaded = false;
    this.http
      .post<{ status: number,message: string, result: string }>(
        `http://${this.host}:3000/api/gethistory`, obj
      )
      .subscribe(response => {
          this.isLoading = false;
          const message = response.message;
          const status = response.status;
          this.viewLayoutResult = [];
          if(status == 200) { 
            this.layoutHistoryResult = [];
            const results: layoutHistory[] = JSON.parse(response.result);
            this.displayedColumns = ['TxId', 'Id', 'Address', 'requestedNOC', 'FAStatus', 'LAStatus', 'ApprovalStatus', 'TimeStamp', 'IsDelete'];
            this.dataSource = new MatTableDataSource(results);  
            this.dataSource.paginator = this.paginator; 
            console.log(this.dataSource);            
            this.historyLoaded = true;                                           
          }  
          console.log("Response message: " + message);
          this.snackBar.open(message, "OK", {
            duration: 0,
          });

      });
    form.resetForm(''); 

  }

  tabChanged(form1: NgForm, form2: NgForm, form3: NgForm){
      form1.resetForm('');
      form2.resetForm('');
      form3.resetForm('');
  }

  
}
