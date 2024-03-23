extern crate pretty_env_logger;

mod telemetry;

use dotenv::dotenv;
use elasticsearch::{http::transport::Transport, Elasticsearch, SearchParts};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use sqlx::postgres::{PgPool, PgPoolOptions};
use std::env;
use warp::{Filter, Rejection, Reply};

#[derive(Debug, Serialize, Deserialize, sqlx::FromRow)]
struct Post {
    id: i64,
    title: String,
    link: String,
    tags: Option<Value>,
}

#[derive(Debug, Serialize, Deserialize, sqlx::FromRow)]
struct SearchQuery {
    query: Option<Value>,
}

async fn get_posts(pool: PgPool) -> Result<impl Reply, Rejection> {
    let span = tracing::info_span!(target: "api.post", "get posts");

    let result = span.in_scope(|| {
        async {
            tracing::info_span!(target: "api.post", parent: &span, "connect db");

            sqlx::query_as::<_, Post>("SELECT id, title, link, tags FROM posts")
            .fetch_all(&pool).await
        }
    }).await;

    match result {
        Ok(posts) => Ok(warp::reply::json(&posts)),
        Err(error) => {
            eprintln!("Error retrieving posts: {:?}", error);
            Err(warp::reject::reject())
        }
    }
}

async fn search_posts(query: SearchQuery, es: Elasticsearch) -> Result<impl Reply, Rejection> {
    let response = es
        .search(SearchParts::Index(&["crawler-posts"]))
        .body(json!({
            "query": {
                "query_string": query
            }
        }))
        .send()
        .await
        .map_err(|error| {
            eprintln!("Elasticsearch request failed: {}", error);
            warp::reject::reject()
        })?;

    let body = response.json::<Value>().await.map_err(|error| {
        eprintln!("Failed to parse Elasticsearch response: {}", error);
        warp::reject()
    })?;

    let mut v = Vec::new();
    for hit in body["hits"]["hits"].as_array().unwrap() {
        let id = hit["_source"]["after"]["id"].clone();

        if v.contains(&id) {
            continue;
        }

        v.push(id);
    }

    Ok(warp::reply::json(&v))
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    if env::var_os("RUST_LOG").is_none() {
        // Set `RUST_LOG=todos=debug` to see debug logs
        env::set_var("RUST_LOG", "posts=info");
    }

    dotenv().ok();
    telemetry::init_tracer();

    let pool: sqlx::Pool<sqlx::Postgres> = PgPoolOptions::new()
        .max_connections(5)
        .connect(&env::var("POSTGRES_DNS").unwrap())
        .await?;

    let transport = Transport::single_node(&env::var("ELASTICSEARCH_DNS").unwrap())?;
    let client = Elasticsearch::new(transport);

    let get_posts = warp::get()
        .and(warp::path("posts"))
        .and(with_db(pool.clone()))
        .and_then(get_posts);

    let search_posts = warp::post()
        .and(warp::path("posts"))
        .and(warp::path("search"))
        .and(json_body())
        .and(with_es(client.clone()))
        .and_then(search_posts);

    let options_route = warp::options().map(warp::reply);

    let cors = warp::cors()
        .allow_any_origin()
        .allow_methods(vec!["GET", "POST", "OPTIONS"])
        .allow_headers(vec!["Content-Type"]);

    let routes = get_posts
        .or(search_posts)
        .or(options_route)
        .with(cors)
        .with(warp::log("posts"));

    warp::serve(routes).run(([0, 0, 0, 0], 3000)).await;

    Ok(())
}

fn with_db(
    pool: PgPool,
) -> impl Filter<Extract = (PgPool,), Error = std::convert::Infallible> + Clone {
    warp::any().map(move || pool.clone())
}

fn with_es(
    es: Elasticsearch,
) -> impl Filter<Extract = (Elasticsearch,), Error = std::convert::Infallible> + Clone {
    warp::any().map(move || es.clone())
}

fn json_body() -> impl Filter<Extract = (SearchQuery,), Error = warp::Rejection> + Clone {
    warp::body::content_length_limit(1024 * 16).and(warp::body::json())
}
