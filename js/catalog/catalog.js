
/* for new 3 column layout*/

function myfcn(p_list) {
    var i;

    var c_out = '';
    var pr_out = '';
    var pu_out = '';


    for(i = 0; i < p_list.length; i++) {
            
        addentry = '<div class="box done">'+'<div class="info">'+'<h5>' + 
        p_list[i].name +
        '<a href="' + p_list[i].url + '">' + 
        '<i class="fa fa-github"></i>'+ '</a>'+
//        '<i class="fa fa-github-alt"></i>'+ '</a>'+
//        '<a href=""><i class="fa fa-user"></i></a>' +

        '<i class="fa fa-star"></i>' + '</h5>' +
        '<p>' + p_list[i].description + '</p>' +
        '</div></div>';

        if(p_list[i].type =="Collector"){
            c_out += addentry;
        }
        else if(p_list[i].type =="Processor"){
            pr_out += addentry;
        }
        else{ // type = publisher
            pu_out += addentry;
        }
    }
    document.getElementById("col_01").innerHTML = c_out;
    document.getElementById("col_02").innerHTML = pr_out;
    document.getElementById("col_03").innerHTML = pu_out;       
}


function myfcn2(p_list2) {
    var i;
    var committed_c_out = '';
    var committed_pr_out = '';
    var committed_pu_out = '';


    for(i = 0; i < p_list2.length; i++) {
        addentry = '<div class="box in-progress">'+'<div class="info">'+'<h5>' + 
        p_list2[i].name +

        '<a href="' + p_list2[i].author_url + '"><i class="fa fa-user"></i></a>' +
        '<i class="fa fa-star-half-full"></i>' + '</h5>' +
        '<p>' + p_list2[i].description + '</p>' +
        '</div></div>';

        if(p_list2[i].type =="Collector"){
            committed_c_out += addentry;
        }
        else if(p_list2[i].type =="Processor"){
            committed_pr_out += addentry;
        }
        else{ // type = publisher
            committed_pu_out += addentry;
        }
    }

    document.getElementById("col_01").innerHTML += committed_c_out;
    document.getElementById("col_02").innerHTML += committed_pr_out;
    document.getElementById("col_03").innerHTML += committed_pu_out; 
}


function myfcn3(p_list3) {
    var i;
    var wish_c_out = '';
    var wish_pr_out = '';
    var wish_pu_out = '';

    for(i = 0; i < p_list3.length; i++) {


        addentry = '<div class="box wish">'+'<div class="info">'+'<h5>' + 
        p_list3[i].name +
        '<i class="fa fa-star-o"></i>'+
        '</h5></div></div>';

        if(p_list3[i].type =="Collector"){
            wish_c_out += addentry;
        }
        else if(p_list3[i].type =="Processor"){
            wish_pr_out += addentry;
        }
        else{ // type = publisher
            wish_pu_out += addentry;
        }

    }
    document.getElementById("col_01").innerHTML += wish_c_out;
    document.getElementById("col_02").innerHTML += wish_pr_out;
    document.getElementById("col_03").innerHTML += wish_pu_out; 
}