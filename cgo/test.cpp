#include "test.h"
#include <iostream>
#include <algorithm>

void test_real(char **names, int nsize, float *features, int fsize)
{

    for (int i = 0; i < nsize; ++i)
        std::cout << names[i] << std::endl;

    std::for_each(features, features + fsize, [](float e) {
        std::cout << e << std::endl;
    });
}

void test(void *names, int nsize, void *features, int fsize)
{
    test_real((char **)names, nsize,
              (float *)features, fsize);
}

std::string testda = "test APp";
const char *app()
{
    return testda.c_str();
}
