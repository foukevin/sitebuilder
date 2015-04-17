SiteBuilder - A static html site generator
==========================================

Copyright (c) 2015 Kévin Delbrayelle

SiteBuilder is an Open Source project covered by the GNU General Public
License version 2.

About SiteBuilder
-----------------------------------------------------------------------

SiteBuilder is a command line tool to generate static html site from a bunch
of markdown formatted articles. It is written in [Go](https://golang.org).

Basic usage
-----------------------------------------------------------------------

Supposing the markdown formatted articles are located into the `./contents`
directory, running the following command will generate a basic set of html
pages:

	$ sitebuilder ./contents

A ccs file can be specified:

	$ sitebuilder --css=style.css ./contents

Templates files for article page can also be specified. Templates use the Go
[HTML Template](http://golang.org/pkg/html/template).

	$ sitebuilder --template=tmpl.html ./contents

Information about the site can be provided using various flags. Site name and
URL:

	$ sitebuilder --name="My site" --url="http://www.mysite.com" ./contents

Author and email:

	$ sitebuilder --author="Kévin Delbrayelle" --email="kevin@mysite.com" ./contents

In addition to the auto-generated archives, an about page link can be provided
in the header.

	$ sitebuilder --about=about.md ./contents
