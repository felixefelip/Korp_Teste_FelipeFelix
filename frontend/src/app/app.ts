import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { FlashMessages } from './shared/flash/flash-messages';
import { Navbar } from './shared/navbar/navbar';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, Navbar, FlashMessages],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class App {}
