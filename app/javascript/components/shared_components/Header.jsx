// BigBlueButton open source conferencing system - http://www.bigbluebutton.org/.
//
// Copyright (c) 2022 BigBlueButton Inc. and by respective authors (see below).
//
// This program is free software; you can redistribute it and/or modify it under the
// terms of the GNU Lesser General Public License as published by the Free Software
// Foundation; either version 3.0 of the License, or (at your option) any later
// version.
//
// Greenlight is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
// PARTICULAR PURPOSE. See the GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License along
// with Greenlight; if not, see <http://www.gnu.org/licenses/>.

import React from "react";
import Container from "react-bootstrap/Container";
import { Navbar } from "react-bootstrap";
import { Link } from "react-router-dom";
import { useAuth } from "../../contexts/auth/AuthProvider";
import Logo from "./Logo";
import NavbarSignedIn from "../home/NavbarSignedIn";
import NavbarNotSignedIn from "../home/NavbarNotSignedIn";
import { siteConfig } from "../../lib/config";

export default function Header() {
  const currentUser = useAuth();

  // let homePath = "/";
  // if (currentUser?.permissions?.CreateRoom === "true") {
  //   homePath = "/rooms";
  // } else if (currentUser?.permissions?.CreateRoom === "false") {
  //   homePath = "/home";
  // }

  return (
    <Navbar collapseOnSelect id="navbar" expand="sm">
      <Container className="ps-0">
        <Navbar.Brand
          as={Link}
          to={siteConfig?.branding?.link}
          className="ps-2 d-flex align-items-center"
        >
          <Logo size="small" />
          <p className="ms-3 mb-0 fs-5 fw-bold">{siteConfig?.branding?.name}</p>
        </Navbar.Brand>
        {currentUser.signed_in && <NavbarSignedIn currentUser={currentUser} />}
      </Container>
    </Navbar>
  );
}
