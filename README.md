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

>GET /{redirectHash}


# Tech/framework Used

atmzr is written using Go with a couple packages, including Mux router.

# Next steps I'd love to do in the future

1. Use Models for the two tables in the database
1. Create a landing page with an HTML template
1. Use Redis to cache the redirect URLs
1. Build a front end with nice UI to shorten URLs
1. Create user login
1. Allow users to manage their links, replace, delete
1. Create an endpoint that returns time series data 
1. Use d3.js to build a nice graph for a link's time series data.
1. Allow users to create custom link shortcodes

# Tests

# Credits