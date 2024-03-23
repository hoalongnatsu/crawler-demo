import React, { useState, useEffect } from "react";
import CssBaseline from "@mui/material/CssBaseline";
import Grid from "@mui/material/Grid";
import Container from "@mui/material/Container";
import { createTheme, ThemeProvider } from "@mui/material/styles";
import Header from "./Components/Header";
import MainFeaturedPost from "./Components/MainFeaturedPost";
import FeaturedPost from "./Components/FeaturedPost";
import Footer from "./Components/Footer";

const sections = [
  { title: "Cloud", url: "#" },
  { title: "DevOps", url: "#" },
];

// TODO remove, this demo shouldn't need to reset the theme.
const Blog = () => {
  const [mainFeaturedPost, setMainFeaturedPost] = useState([]);
  const [featuredPosts, setFeaturedPosts] = useState([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchData, setSearchData] = useState("");

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch(`${process.env.REACT_APP_API_URL}/posts`);
        const data = await response.json();
        setMainFeaturedPost(data.shift());
        setFeaturedPosts(data);
      } catch (error) {
        console.error("Error fetching data:", error);
      }
    };

    fetchData();
  }, []);

  const onSearch = async () => {
    try {
      const response = await fetch(
        `${process.env.REACT_APP_API_URL}/posts/search`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            query: searchQuery,
          }),
        }
      );
      const searchData = await response.json();
      setSearchData(searchData);
    } catch (error) {
      console.error("Error searching:", error);
    }
  };

  return (
    <ThemeProvider theme={createTheme()}>
      <CssBaseline />
      <Container maxWidth="lg">
        <Header
          title="Crawler System"
          sections={sections}
          onSearch={onSearch}
          setSearchQuery={setSearchQuery}
        />
        <main>
          <MainFeaturedPost post={mainFeaturedPost} />
          <Grid container spacing={4}>
            {featuredPosts.map((post) => {
              if (searchData.length === 0) {
                return <FeaturedPost key={post.title} post={post} />;
              } else if (searchData.includes(post.id)) {
                return <FeaturedPost key={post.title} post={post} />;
              }

              return null;
            })}
          </Grid>
        </main>
      </Container>
      <Footer title="Footer" description="Demo Crawler System" />
    </ThemeProvider>
  );
};

export default Blog;
