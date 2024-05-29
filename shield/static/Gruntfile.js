module.exports = function (grunt) {
  grunt.initConfig({
    packgen: grunt.file.readJSON("package.json"),
    uglify: {
      target: {
        files: [
          {
            src: "js/**/*.js",
            dest: "build/min.js",
          },
        ],
      },
    },
    cssmin: {
      target: {
        files: [
          {
            src: "assets/css/**/*.css",
            dest: "build/min.css",
          },
        ],
      },
    },
    htmlmin: {
      options: {
        collapseWhitespace: true,
      },
      target: {
        files: [
          {
            src: "src/**/*.html",
            dest: "build/min.html",
          },
        ],
      },
    },
  });

  grunt.loadNpmTasks("grunt-contrib-uglify");
  grunt.loadNpmTasks("grunt-contrib-cssmin");
  grunt.loadNpmTasks("grunt-contrib-htmlmin");

  grunt.registerTask("default", ["uglify", "cssmin", "htmlmin"]);
};
