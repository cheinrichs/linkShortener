# atmzr
atmzr (atomizer) is a URL shortening API, developed by Cooper Heinrichs

Currently live online at: http://atmzr.herokuapp.com/

 
# API Reference

Create a shortened link

>POST /createLink
> 
>{url: "http://www.example.com"}

Get Link Statistics

>GET /linkStatistics/{redirectHash}

Redirect using a shortened link

>Get /{redirectHash}


# Tech/framework Used

atmzr is written using Go with a couple packages, including Mux router.

# Features

# Code Example

# Installation

# Tests

# Credits